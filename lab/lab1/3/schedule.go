package mapreduce

import "fmt"

// schedule starts and waits for all tasks in the given phase (Map or Reduce).
func (mr *Master) schedule(phase jobPhase) {
	var ntasks int
	var nios int // number of inputs (for reduce) or outputs (for map)

	var taskDone =make(chan int)
	var watcherDone =make(chan int)
	switch phase {
		case mapPhase:
			ntasks = len(mr.files)  //总共的任务数量
			nios = mr.nReduce		//reduce数量
		case reducePhase:
			ntasks = mr.nReduce		//总共的reduce任务
			nios = len(mr.files)	//map数量
	}

	var workerSize =len(mr.workers)
	var tasks=make([] int,ntasks)

	fmt.Printf("Schedule: %v %v tasks (%d I/Os)\n", ntasks, phase, nios)
	fmt.Printf("[debug]Schedule: mapPhase,worker size:%d\n", len(mr.workers))

	
	//开启线程，不断监听是否有新的worker加入
	go func(){
		for{
			select{
				case worker:=<-mr.registerChannel:
					fmt.Printf("[debug]register worker:%s\n",worker);
					go workerDoTask(worker,phase,mr,tasks,taskDone,nios)
				case <-watcherDone:
					break
			}
		}
	}()

	for i:=0;i<workerSize;i++{
		go workerDoTask(mr.workers[i],phase,mr,tasks,taskDone,nios)
	}

	<-taskDone
	watcherDone<-1
	fmt.Printf("Schedule: %v phase done\n", phase)
}

func  workerDoTask(worker string,phase jobPhase,mr *Master,tasks []int,ch chan int,nios int) {
	
			for{
				taskNumber:=-1
				done:=true
				mr.Lock();
				for i:=0;i<len(tasks);i++{
					//fmt.Println("ntask[i]=%d",tasks[i])
					if tasks[i]==0 { //未执行
						taskNumber=i
						tasks[i]=1 //执行中
						done=false
						break
					}
					if tasks[i]!=2 {

						done=false
					}
				}
				mr.Unlock();
				if done {
					fmt.Println("%s Taskdone bye done done done",phase)
					ch<-1
					return
				}
				if taskNumber==-1{
					continue
				}
				//获取任务编号，创建任务参数，RPC调用worker执行
				fmt.Printf("Worker.DoTask iseeyou : %d\n",taskNumber)
				taskArg:=DoTaskArgs{mr.jobName,mr.files[taskNumber],phase,taskNumber,nios}
				ok := call(worker,"Worker.DoTask",taskArg,new(struct{}));
				if ok==true {
					tasks[taskNumber]=2
					fmt.Printf("Worker.DoTask: %v phase ok\n", phase)
				}else{
					tasks[taskNumber]=0
					fmt.Printf("Worker.DoTask: %v phase failed\n", phase)
				}
			}
}