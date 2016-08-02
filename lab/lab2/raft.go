package raft

//
// this is an outline of the API that raft must expose to
// the service (or tester). see comments below for
// each of these functions for more details.
//
// rf = Make(...)
//   create a new Raft server.
// rf.Start(command interface{}) (index, term, isleader)
//   start agreement on a new log entry
// rf.GetState() (term, isLeader)
//   ask a Raft for its current term, and whether it thinks it is leader
// ApplyMsg
//   each time a new entry is committed to the log, each Raft peer
//   should send an ApplyMsg to the service (or tester)
//   in the same server.
//

import "sync"
import "labrpc"
import "strconv"
import "time"
// import "fmt"
import "math/rand"
import "strings"

import "bytes"
import "encoding/gob"



//
// as each Raft peer becomes aware that successive log entries are
// committed, the peer should send an ApplyMsg to the service (or
// tester) on the same server, via the applyCh passed to Make().
//
type ApplyMsg struct {
	Index       int
	Command     interface{}
	UseSnapshot bool   // ignore for lab2; only used in lab3
	Snapshot    []byte // ignore for lab2; only used in lab3
}

type LogEntry struct{
	cmd int
	committed bool //是否提交
}
//
// A Go object implementing a single Raft peer.
//
type Raft struct {
	mu        sync.Mutex
	peers     []*labrpc.ClientEnd
	persister *Persister
	me        int // index into peers[]

	// Your data here.
	// Look at the paper's Figure 2 for a description of what
	// state a Raft server must maintain.
	// commitOver chan bool
	asking bool
	currentTerm int //latest term server has seen(初始化为0，单调递增)
	votedFor string //received vote in currentTerm
	serverId string //server id
	logs     []int
	count 	int
	chans     []chan int
	hbCh     chan string
	role	int  // 0: leader 1 :candidate 2 :follower
	heartBeat bool
	applyCh chan ApplyMsg
	hbReply int
	killChan chan int
	receiveCommand chan int
	lastCommitIndex int //last commitedIndex
	//just for leader
	nextIndex []int //保存每台服务器的 下次发送的log entry的index(初始化为leader的最后一个log的index+1)
	matchIndex []int //保存每台服务器的 最高的复制过的log entry的index（初始化为0，单调递增）

}

// return currentTerm and whether this server
// believes it is the leader.
func (rf *Raft) GetState() (int, bool) {

	var term int
	var isleader bool
	// Your code here.

	term=rf.currentTerm
	if rf.role==0{
		isleader=true
	}else{
		isleader=false;
	}

	return term, isleader
}

//
// save Raft's persistent state to stable storage,
// where it can later be retrieved after a crash and restart.
// see paper's Figure 2 for a description of what should be persistent.
//
func (rf *Raft) persist() {
	// Your code here.
	// Example:
	rf.DPrintf("[persist].... %d",rf.count)
	w := new(bytes.Buffer)
	e := gob.NewEncoder(w)
	e.Encode(rf.count)
	e.Encode(rf.currentTerm)
	e.Encode(rf.votedFor)
	e.Encode(rf.lastCommitIndex)
	e.Encode(rf.logs)
	data := w.Bytes()
	rf.persister.SaveRaftState(data)
	
}

//
// restore previously persisted state.
//
func (rf *Raft) readPersist(data []byte) {
	// Your code here.
	// Example:
	rf.DPrintf("[readPersist].... %s",string(data))
	r := bytes.NewBuffer(data)
	d := gob.NewDecoder(r)
	d.Decode(&rf.count)
	d.Decode(&rf.currentTerm)
	d.Decode(&rf.votedFor)
	d.Decode(&rf.lastCommitIndex)
	d.Decode(&rf.logs)

	// rf.DPrintf("...... log[1]=%d",rf.logs[1])
}




//
// example RequestVote RPC arguments structure.
//
type RequestVoteArgs struct {
	// Your data here.
	Term int //
	CandidateId  string
	LastLogIndex int
	LastLogTerm int
}

//
// example RequestVote RPC reply structure.
//
type RequestVoteReply struct {
	// Your data here.
	Term int
	Result bool
}

//
// example RequestVote RPC handler.
//
func (rf *Raft) RequestVote(args RequestVoteArgs, reply *RequestVoteReply) {
	// Your code here.
	begin:=rf.currentTerm

	rf.DPrintf("[RequestVote]begin %s vote to-->%s count=%d & LastLogIndex=%d",rf.serverId,args.CandidateId,rf.count,args.LastLogIndex)
	if rf.count>args.LastLogIndex{
		reply.Result=false
		rf.DPrintf(">[RequestVote]%s <-- %s  %d & %d= %s", rf.serverId,args.CandidateId,begin,args.Term,reply.Result)
		return
	}
	if rf.currentTerm<args.Term{//当前的term小于等于RPC的term
		rf.currentTerm=args.Term
		rf.votedFor=args.CandidateId
		rf.role=2
		reply.Result=true
	}else if rf.currentTerm>args.Term{//收到了小于等于的term，拒绝
		reply.Result=false
		reply.Term=rf.currentTerm
	}else{
		// reply.Result=true
		rf.DPrintf("[RequestVote] %s with %s",args.CandidateId,rf.votedFor)
		if strings.Compare(args.CandidateId,rf.votedFor)==0{
			reply.Result=true
			rf.role=2
		}else{
			reply.Result=false
		}

	}
	
	rf.DPrintf("[RequestVote]%s <-- %s  %d & %d= %s", rf.serverId,args.CandidateId,begin,args.Term,reply.Result)
}

//
// example code to send a RequestVote RPC to a server.
// server is the index of the target server in rf.peers[].
// expects RPC arguments in args.
// fills in *reply with RPC reply, so caller should
// pass &reply.
// the types of the args and reply passed to Call() must be
// the same as the types of the arguments declared in the
// handler function (including whether they are pointers).
//
// returns true if labrpc says the RPC was delivered.
//
// if you're having trouble getting RPC to work, check that you've
// capitalized all field names in structs passed over RPC, and
// that the caller passes the address of the reply struct with &, not
// the struct itself.
//
func (rf *Raft) sendRequestVote(server int, args RequestVoteArgs, reply *RequestVoteReply) bool {
	//fmt.Println("rf.peers[server].:",rf.peers[server]);
	ok := rf.peers[server].Call("Raft.RequestVote", args, reply)
	return ok
}



type AppendEnriesArgs struct {
	IsHeartBeat bool	//是否是心跳
	FromServer string  //leader
	Term int            //leader的trem
	Msg  ApplyMsg       //cmd
	PrevLog int //上一个cmd
	Commit bool // ask or commit
	IsLeader int //0 is leader
	Index int    //添加的index
}

type AppendEnriesReply struct {
	Result bool
	Term int
}

func (rf *Raft) AppendEnries(args AppendEnriesArgs, reply *AppendEnriesReply) {
	if(args.IsHeartBeat&&rf.currentTerm<=args.Term){//收到的是心跳且当前服务器的term小于等于RPC的term
		rf.hbCh<-args.FromServer
		rf.role=2 //follower
		rf.heartBeat=true
		rf.currentTerm=args.Term//如果当前Raft服务器的Term小于RPC的term,则设置为leader的term
	}
	if !args.IsHeartBeat{// append entries
		if(rf.currentTerm<=args.Term){//当前的term小于RPC的term
			rf.currentTerm=args.Term
			// if args.PrevLog==rf.logs[args.Msg.Index-1]{//若RPC的prev与本服务器logs相等，则可以添加
				if args.Commit{	//提交

					if args.PrevLog!=rf.logs[args.Msg.Index-1]{
						reply.Result=false
						reply.Term=-1
						// for k:=1;k<10;k++{
						// 	rf.DPrintf("[AppendEnries] k[%d]=%d ",k,rf.logs[k])
						// }
						rf.DPrintf("[AppendEnries]sorry!,prev not equals! args.PrevLog :%d & rf.logs[%d] %d",args.PrevLog,args.Msg.Index-1,rf.logs[args.Msg.Index-1])
					}else{
						rf.applyCh<-args.Msg //发送过去才算commit
						if 0!=args.IsLeader{
							rf.count++
							if rf.lastCommitIndex<args.Index{
								rf.lastCommitIndex=args.Index
							}
						}
						
						rf.DPrintln("[AppendEnries]received append entries!",args.Index,"  ",args.Msg.Command)
						rf.DPrintf("[AppendEnries]%s received append entries! is leader(0):%d rf.logs[%d]=%d",rf.serverId,args.IsLeader,args.Index,args.Msg.Command)
						rf.logs[args.Index]=args.Msg.Command.(int)
						rf.persist()
						reply.Result=true
					}
				}else{
					rf.DPrintln("[AppendEnries]leader just ask for ",args.Msg.Command)
					reply.Result=true
				}
					
			// }else{
			// 	fmt.Println("args.PrevLog ",args.PrevLog," rf.logs[args.Msg.Index-1]",rf.logs[args.Msg.Index-1])
			// 	reply.Result=false
			// }
			
		}else{//当前的term大于PRC的term
			reply.Term=rf.currentTerm
			reply.Result=false
		}
		
	}
}

func (rf *Raft) sendAppendEntries(server int, args AppendEnriesArgs, reply *AppendEnriesReply) bool {
	//
	// rf.DPrintf("RPC sendAppendEntries");
	ok := rf.peers[server].Call("Raft.AppendEnries", args, reply)
	return ok
}
//
// the service using Raft (e.g. a k/v server) wants to start
// agreement on the next command to be appended to Raft's log. if this
// server isn't the leader, returns false. otherwise start the
// agreement and return immediately. there is no guarantee that this
// command will ever be committed to the Raft log, since the leader
// may fail or lose an election.
//
// the first return value is the index that the command will appear at
// if it's ever committed. the second return value is the current
// term. the third return value is true if this server believes it is
// the leader.
//
func (rf *Raft) Start(command interface{}) (int, int, bool) {
	index := -1
	term := -1
	isLeader := true
	// pCount :=len(rf.peers)
	if rf.role!=0{//isn't leader returns false
		isLeader=false
	}else{
		
			rf.mu.Lock();
			index=rf.count;
			rf.count++
			rf.DPrintf("[Start] preaddto queue, cmd=%d  index=%d",command,index)
			//go rf.sendAppendEntriesRPC(rf.me,pCount,command.(int),index)
			rf.receiveCommand<-command.(int) //加入channel缓存
			rf.mu.Unlock();
	}
	term=rf.currentTerm
	return index, term, isLeader
}


//发送log entry RPC
func (rf *Raft) sendAppendEntriesRPC(me int,pCount int) {
	//若rf已经不是leader了 需要停止处理
	for cmd:=range rf.receiveCommand{
	if rf.role!=0{
		continue
	}
	index:=rf.lastCommitIndex+1
	cmtCount:=1 //保存同意的数量，超过一半可以提交
	commitChan :=make(chan bool) 
	timeout :=time.After(time.Second*2)
	canCommit:=false

	rf.DPrintln("[Start]***","serverId:",rf.serverId," index:",index," cmd: ",cmd)
	// sum:=0b
	if index>1{
		if(rf.logs[index-1]!=0){//由于leader肯定包含所有的logentry,上一个不为0表示已经完成，不需要等待
			rf.DPrintln("[sendAppendEntriesRPC]not need waiting... rf.logs[index-1] ",rf.logs[index-1])
		}else{
			//等待上一个commit
			// rf.barried++
			// rf.DPrintln("[sendAppendEntriesRPC]create rf.chans[] ",index-1)
			// rf.chans[index-1]=make(chan int)
			// rs:=<-rf.chans[index-1]
			// rf.DPrintf("[sendAppendEntriesRPC]get chans[%d]=%d ,begin add!",index-1,rs)
			// rf.barried--
			// if rs==0{	//等待上一个add超时，不等了
			// 	rf.count--
			// 	continue
			// }
		}
	}

	for i:=0;i<pCount;i++{
			//发送多个ask
			go func(i int,cmd interface{}){
				//向index发出ask请求
				if i==me{
					return
				}
				prev:=rf.logs[index-1] 
				msg:=ApplyMsg{Index:index,Command:cmd}
				req:=AppendEnriesArgs{Index:index,IsHeartBeat:false,FromServer:rf.serverId,Term:rf.currentTerm,Msg:msg,PrevLog:prev,Commit:false,IsLeader:i-me}
				rep:=new(AppendEnriesReply)
				rf.DPrintf("[sendAppendEntriesRPC] ask server%d.....index=%d cmd=%d",i,index,cmd)
				ok:=rf.sendAppendEntries(i,req,rep)

				if ok{
					if rep.Result{//成功了
						rf.mu.Lock()
						cmtCount++
						rf.mu.Unlock()

						if cmtCount>=(pCount+1)/2 {
							commitChan<-true
						}
					}
				}
			}(i,cmd)
				
	}

	// 因为如果时超时的话，commitChan收不到了啊
	select{
		case canCommit=<-commitChan:
			
			if !canCommit{//没有通过commit，减去1,但是由于没有for，收到timeout就over了，timeout不会走这里
				// rf.count-- //may be need niled chan
				// fmt.Println("becauseof canCommit is false,set chans",index,rf.chans[index]!=nil)
				// if rf.chans[index]!=nil{
				// 	rf.chans[index]<-1
				// 	rf.chans[index]=nil
				// }
				return
			}else{//通过commit，去提交

				rf.DPrintf("[sendAppendEntriesRPC]OK!I CAN COMMIT! index:%d cmd:%d",index, cmd)
				rf.logs[index]=cmd;
				rf.lastCommitIndex++
				rf.persist()
				retCnt:=1
				for i:=0;i<pCount;i++{
					go func(i int,index int,cmd interface{}){

						nowtime:=time.Now().UnixNano()
						if i==me{	
							rf.applyCh<-ApplyMsg{Index:index,Command:cmd}
							//放这里啊。。。底下不行 如有失败的（TestFailNoAgree过不了）
							// rf.DPrintln("[sendAppendEntriesRPC]commit self over!rf.chans[",index,"]=",rf.chans[index]==nil)
							// if rf.chans[index]!=nil{
							// 	rf.chans[index]<-1
							// 	rf.chans[index]=nil
							// }
							return
						}
						for{//失败了需要一直重试 丧心病狂啊,但是当被后续commit补充后则可以停在
						myIndex:=rf.nextIndex[i]
						//首先应该校验prev的数据，若myIndex小于index，表明需要先获取数据
						//myIndex肯定是小于且等于index的

						rf.DPrintf("[sendAppendEntriesRPC]begin commit to server%d, compare nextIndex=%d and index=%d",i,myIndex,index)

						for myIndex<=index{

							prev:=rf.logs[myIndex-1] //leader的前一个数据
							myCmd:=rf.logs[myIndex]	//leader的当前数据
							
							//不能一直发请求，比如leader重新选举后nextIndex变化了，
							//但是当前的nextIndex数组是旧leader的，不一定新
							//如果当前rf的count大于了index，则说明可以结束这个线程了
							//后续的index commit会补充，如index到11了，count为9，index5则不需要继续了

							//新的问题，如果并发阻塞了rf.count为6个，其实此时只commit了1个，这时就不继续了
							//这是不对的，应该计算commit的数量

							//leader上一次积累了13个，count是13，当前index是1, 第一个就不继续了咋整，用lastCommitIndex
							rf.DPrintln("[sendAppendEntriesRPC]",nowtime,"compare  lastCommitIndex=",rf.lastCommitIndex," index=",index)
							if rf.lastCommitIndex>index{ //最后提交的是10 10以下才继续
								return
							}

							//发送请求，修改一下follower的数据
							rf.DPrintln("[sendAppendEntriesRPC]",nowtime,"inconsistent> append entries to ",i," myIndex ",myIndex," cmd: ",myCmd," prev ",prev," index",index)
							msg:=ApplyMsg{Index:myIndex,Command:myCmd}
							req:=AppendEnriesArgs{Index:myIndex,IsHeartBeat:false,FromServer:rf.serverId,Term:rf.currentTerm,Msg:msg,PrevLog:prev,Commit:true,IsLeader:i-me}
							rep:=new(AppendEnriesReply)
							rf.DPrintf("[sendAppendEntriesRPC] %d commit.....",index)
							ok:=rf.sendAppendEntries(i,req,rep)
							if ok{
								if rep.Result{//成功添加了
									rf.nextIndex[i]++ //设置nextIndex
									myIndex=rf.nextIndex[i]
									
									//若是leader自己
									retCnt++

									if retCnt>=(pCount+1)/2{//已经ok啦
										rf.DPrintln("[sendAppendEntriesRPC]commit over!rf.chans[",index,"]=",rf.chans[index]==nil)
										if rf.chans[index]!=nil{
											rf.chans[index]<-1
											rf.chans[index]=nil
										}
									}
									if myIndex>index{
										return
									}
								}else{//添加失败
									if rep.Term>0{//需要更新term
										rf.currentTerm=rep.Term
										rf.DPrintln("[sendAppendEntriesRPC]",nowtime," rf.currentTerm................")
									}else if rep.Term<0{//prev不一致，继续向后
										rf.DPrintln("[sendAppendEntriesRPC] nextIndex need -- ",i," index ",myIndex)
										rf.nextIndex[i]--;
										myIndex=rf.nextIndex[i]
									}
								}
							}else{
								// return //如果在此处结束的话，重连之后不会继续
							}
						}
						return
						}//for
					}(i,index,cmd)
				}
				continue
			}
		case <-timeout:
			if !canCommit{ //超时了，还没有收到可以提交
				rf.DPrintf("[sendAppendEntriesRPC]>>>timeout...cannot commit  index=%d  cmd=%d cmtCount=%d",index,cmd,cmtCount)
		
				if cmtCount>=(pCount+1)/2 {
					commitChan<-true
				}else{
					// commitChan<-false
					rf.timeout(index)
				}
			}
	}
	}
}

func (rf *Raft) timeout(index int){
	rf.count-- //may be need niled chan
	rf.DPrintln("[sendAppendEntriesRPC]becauseof canCommit is false,set chans",index,rf.chans[index]!=nil)
	if rf.chans[index]!=nil{
		rf.chans[index]<-1
		rf.chans[index]=nil
	}
}
//
// the tester calls Kill() when a Raft instance won't
// be needed again. you are not required to do anything
// in Kill(), but it might be convenient to (for example)
// turn off debug output from this instance.
//
func (rf *Raft) Kill() {
	// Your code here, if desired.
	rf.role=2
	rf.killChan<-1
}

//
// the service or tester wants to create a Raft server. the ports
// of all the Raft servers (including this one) are in peers[]. this
// server's port is peers[me]. all the servers' peers[] arrays
// have the same order. persister is a place for this server to
// save its persistent state, and also initially holds the most
// recent saved state, if any. applyCh is a channel on which the
// tester or service expects Raft to send ApplyMsg messages.
// Make() must return quickly, so it should start goroutines
// for any long-running work.
//
func Make(peers []*labrpc.ClientEnd, me int,
	persister *Persister, applyCh chan ApplyMsg) *Raft {
	rf := &Raft{}
	rf.peers = peers
	rf.persister = persister
	rf.me = me

	// Your initialization code here.
	pCount :=len(peers)

	// rf.commitOver=make(chan bool)
	rf.logs = make([]int,1000)
	rf.chans=make([]chan int,1000)
	rf.count=1
	rf.hbReply=0
	// rf.loghoder = make([]int,1)
	rf.serverId="server"+strconv.Itoa(me)
	rf.nextIndex = make([]int,pCount)
	rf.currentTerm=0
	rf.hbCh =make(chan string)
	rf.role=2 //默认是follower
	rf.heartBeat=false	//是否收到心跳
	rf.applyCh=applyCh
	rf.asking=false
	rf.lastCommitIndex=0
	rf.receiveCommand=make(chan int,100)
	rand.Seed(int64(time.Now().Nanosecond()))
	rf.killChan=make(chan int)

	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())

	//创建Raft服务器，开启线程，处理各种
	go func() {
		//获取随机的election timeout
		var max time.Duration=time.Duration(rand.Intn(100)+300)
		chaner :=time.Tick(time.Millisecond*max)
		rf.DPrintln("[Make]",me," server ",rf.serverId,"begin ",max);
		for{
			select{
				case <-chaner :
					//election timeout
					//问问有没有收到心跳，若有保持follower，若无开始选举
					//如果是0 不需要根据心跳选举 如果是1说明正在选举
					if(!rf.heartBeat){//如果未收到信息
						if rf.role!=0{//开始选举
							go rf.elect(me,pCount)
						}
					}else{//election time内收到心跳了,重置
						rf.heartBeat=false
					}
				case frm:=<-rf.hbCh://收到心跳
					rf.DPrintln("[Make.heartbeat]",rf.serverId," get RPC heartbeat from ",frm);
					rf.heartBeat=true
					rf.role=2
				case <-rf.killChan:
					return
			}
		}
	}()

	
	go rf.sendAppendEntriesRPC(me,len(rf.peers))

	return rf
}



//开始选举
func (rf *Raft) elect(me int,pCount int){
	//如果已经是candidate 不需要再次选举了吧,需要比较term吧
	rf.DPrintln("[elect]i want to take part in election! ",rf.serverId,rf.heartBeat,rf.role);
	rf.role=1 //candidate
	rf.currentTerm=rf.currentTerm+1
	voteCnt:=1 //vote for self
	rf.votedFor=rf.serverId//vote for self
	sum :=0
	for i:=0;i<pCount;i++{
		if i==me{//如果是自己就不发生RPC请求了
			continue;
		}
		//并发发送vote request
		go func(i int,rf *Raft){
			rf.DPrintln("[elect]",rf.serverId," send vote request to ",i," term: ",rf.currentTerm," voteCnt:",voteCnt)
			beg:=time.Now().UnixNano()/1000
			rf.mu.Lock();
			sum++ // send request count +1
			rf.mu.Unlock();
			req:=RequestVoteArgs{
				rf.currentTerm,
				rf.serverId,rf.count,0}
			rep:=new(RequestVoteReply)
			ret:=rf.sendRequestVote(i,req,rep)
		
			// fmt.Println("ret......... !",i,rf.serverId,ret)
			if !ret{	//调用失败，说明离线了
				rf.DPrintln("[elect]sorry error!",time.Now().UnixNano()/1000-beg," term: ",rf.currentTerm) 
				rf.DPrintln("[elect]voteCnt=",voteCnt)
				return
			}else{
				rf.DPrintln("[elect]vote success!",time.Now().UnixNano()/1000-beg," term: ",rf.currentTerm)
				//投票了
				if rep.Result{
					rf.mu.Lock();
					voteCnt++
					rf.mu.Unlock();
				}else{
					rf.currentTerm=rep.Term
				}
			}
			//若为candidate且投票大于一半且所有请求均已经发出
			if rf.role==1&&voteCnt>=(pCount+1)/2&&(sum==pCount-1) {
				rf.role=0//you are leader
				//发送心跳
				rf.DPrintln("[elect]i am the leader!",rf.serverId)
				rf.DPrintln("[elect]send RPC over! voteCnt=",voteCnt);
				go rf.sendHeartBeat(me,pCount)
				//初始化nextIndex数组
				for n:=0;n<pCount;n++{
					rf.nextIndex[n]=rf.count//len(rf.logs);
				}
			}else{

			}
		}(i,rf)	
	}

}




func (rf *Raft) sendHeartBeatRPC(me int,pCount int) {

	for i:=0;i<pCount;i++{
		if i==me{//如果是自己就不发生RPC请求了
			continue;
		}
		go func(i int){
			rf.DPrintln("[elect]",rf.serverId," send heartbeat to",i)
			req:=AppendEnriesArgs{IsHeartBeat:true,FromServer:rf.serverId,Term:rf.currentTerm,IsLeader:i-me}
			rep:=new(AppendEnriesReply)
			// rf.DPrintln("heartBeat.....")
			rf.sendAppendEntries(i,req,rep)
			rf.hbReply++
			// rf.DPrintln("[elect]",rf.serverId,"send heartbeat success!")
		}(i)
	}
	
}

//发送心跳
func (rf *Raft) sendHeartBeat(me int,pCount int) {
	//由于down后发送的RPC返回台湾，加入超时时间，定期发送
	var max time.Duration=time.Duration(rand.Intn(100)+250)
	chaner :=time.Tick(time.Millisecond*max)
	
	//开启线程，不断的发送心跳
	go func(){

		//先立刻发送心跳包
		rf.sendHeartBeatRPC(me,pCount)

		for{
			select{
				case <-chaner :
					if(rf.hbReply==0){
						rf.DPrintln("[sendHeartBeat]",rf.serverId," because of  heaertbeat become follower!")
						rf.role=2//发送心跳 未收到任何响应 变为follower
						for i:=1;i<len(rf.chans);i++{ //close chans
							if rf.chans[i]!=nil{
								rf.count--
								rf.chans[i]<-0
							}
						}

					}else{
						rf.hbReply=0
					}
					if rf.role!=0{ //不为leader 退出
						rf.DPrintln("[sendHeartBeat]","i am not a leader,quit!",rf.serverId)
						return
					}
					rf.sendHeartBeatRPC(me,pCount)
			}
		}

	}()
}