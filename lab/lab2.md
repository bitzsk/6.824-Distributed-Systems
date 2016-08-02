# 实验二：Raft

---

##介绍

从这个实验开始，我们将开始一系列关于创建可容错的key/value存储系统的实验。在实验二中，你需要实现Raft，一个复制状态机协议。下个实验中，你需要在Raft的基础上创建K/V服务，然后优化系统提高性能，最后实现分布式事务。

一个复制服务（如，key/value数据库）可以使用Raft帮助其管理副本服务器。副本服务器是为了整个系统在某些服务器发生故障或网络中断时还可以继续提供服务。实验的挑战在于，由于这些故障，副本之间不会一直保持相同的数据.Raft可以帮助服务器找出正确的数据。

Raft的基本的作用是实现一个复制状态机。其将client的请求组织成一个序列，称之为log，确保所有副本的log内容统一。每个副本以相同顺序执行log中的clietn请求，这样多个副本有相同的服务状态。如果服务发生故障，但稍后回复。Raft会将其状态更新为最新。Raft会不断的处理请求，只要多数服务器还在运行并可以保持相互联系。如果超过一半服务器停止运行，Raft也会停止，但是当多数服务器恢复后可以很快恢复。

在本实验中，需要用Go语言实现Raft算法，多个Raft实例通过RPC交互维持复制状态。Raft接口需要支持不确定的，并通过index进行编号的command序列，称为log entries。给定index的log entries最终将会被提交。Rast应该发生log entry到其他多数服务器上执行。

    注意：RPC只能用来在不同的Raft实例之间进行交互。例如，不同Raft实例不允许共享Go变量。你的实现程序不应该使用文件。
    
在这个实验中，你需要实现论文中大多数的Raft算法设计，包含保存持久化状态，节点发送故障重启后读取持久化状态。你不需要实现集群成员关系的更改和log压缩/快照。

你可以从论文和课程笔记中获取相关知识点，也可以从[动画演示](http://thesecretlivesofdata.com/raft/)中学到Raft。另外，可以看看Paxos、Chubby、Paxos Made Live, Spanner, Zookeeper, Harp, Viewstamped Replication, and Bolosky 等待。

    Paxos：Lamport提出的一种基于消息传递的一致性算法
    
    Chubby：Google Chubby通过锁原语为其他系统实现更高级的服务，比如组成员、域名服务和leader选举等等。采用Leslie Lamport提出的paxos算法来实现可靠容错，这是业界关于paxos第一个完整可行的实现
    
    Paxos Made Live：最初实现Paxos碰到的一系列问题及解决方案，是一篇全面讲解分布式系统工程实践的文章。
     
    Spanner：Spanner是谷歌公司研发的、可扩展的、多版本、全球分布式、同步复制数据库。它是第一个把数据分布在全球范围内的系统，并且支持外部一致性的分布式事务。
    
    ZooKeeper：这是一个分布式的，开放源码的分布式应用程序协调服务，是Google的Chubby一个开源的实现，是Hadoop和Hbase的重要组件。
    
    Harp：基于VFS接口的可复制Unix文件系统，提供了高可用性和可依赖的文件存储。不管是并发还是发生故障，都可以保证文件操作时原子性。
    
    Viewstamped Replication:一个支持高可用性的分布式复制算法
    
    Bolosky：http://static.usenix.org/event/nsdi11/tech/full_papers/Bolosky.pdf
    
* 开始之前，虽然实现Raft的代码不会很多，但是使之能够正常工作是很有挑战的。算法和代码都比较棘手，需要考很多边角的case。当其中一个用例失败，可能需要花费一些时间理解程序出错的场景以及如何修复程序。
* 在开始实验之前，请先阅读论文和课程笔记。你实现的程序应该很解决论文的描述，因为这样才能通过测试用例。图2描述的伪代码对实现算法来说很有用。


##开始
使用`git pull`获取最新的实验程序。我们在`src/raft`中提供了程序框架和测试用例，`src/labrpc`中提供了一个简单的类RPC系统。
使用下列命令可以让程序开始运行：
```
$ setup ggo_v1.5
$ cd ~/6.824
$ git pull
...
$ cd src/raft
$ GOPATH=~/6.824
$ export GOPATH
$ go test
Test: initial election ...
--- FAIL: TestInitialElection (5.03s)
	config.go:270: expected one leader, got 0
Test: election after network failure ...
--- FAIL: TestReElection (5.03s)
	config.go:270: expected one leader, got 0
...
$
```
若你的程序通过所有用例(`src/raft`)，你会看到
```
$ go test
Test: initial election ...
  ... Passed
Test: election after network failure ...
  ... Passed
...
PASS
ok  	raft	162.413s
```

##任务

你需要通过在`raft/raft.go`添加你的代码实现Raft。在这个文件中，你会发现一些框架代码，还有一些如何发送和接收RPC的例子，还有如何保存/恢复持久化状态的例子。
```
// create a new Raft server instance:
rf := Make(peers, me, persister, applyCh)

// start agreement on a new log entry:
rf.Start(command interface{}) (index, term, isleader)

// ask a Raft for its current term, and whether it thinks it is leader
rf.GetState() (term, isLeader)

// each time a new entry is committed to the log, each Raft peer
// should send an ApplyMsg to the service (or tester).
type ApplyMsg
```

`Make(peers,me,…)`用来创建一个Raft节点。参数peers是已完成的RPC连接的数组，每个连接到一个Raft节点。参数me是这个节点在peers中的index。
`Start(command)`Raft启动进程，追加command到复制log。这个方法应该立即返回，而不会等待操作完成。
服务希望你的程序发送`ApplyMsg`到applyCh参数到'Make()'

你的Raft节点应该使用我们提供的labrpc包来交换RPC。labrpc模仿了go语言的[rpc库](https://golang.org/pkg/net/rpc/),内部使用Go channel机制而不是socket。`raft.go`包含一些发送RPC（`sendRequestVote()`）和处理RPC请求(`RequestVote()`)的示例代码。

实现leader选举和心跳（空的AppendEntries请求）。这对于选举并保持单个leader是足够的。一旦你完成了这些工作，可以通过"go test"前两个测试用例。

* 在`raft.go`中，将你需要的所有状态添加到`Raft`结构体中。论文的图2提供了引导。你也需要定义一个结构体来保存每个log entry的信息。记住任何需要通过RPC发送的结构体的属性都需要以大写字母打头。
* 你应该首先实现Raft的leader选举。填充`RequestVoteArgs`和`RequestVoteReply`结构体，修改`Make()`当给定时间未收到其他节点消息时创建一个底层goroutine启动选举（通过发送`RequestVote` RPCs）。为了使选举能够工作，你需要实现RequestVote()RPC处理以便可以投票。
* 为了实现心跳，你需要定义`AppendEntries` RPC结构体（虽然你可能还不需要所有参数），leader周期性发送。你还需要编写`AppendEntries`处理方法重置`election timeout`以便leader选举完成后其他服务器不会继续作为leader。
* 确保不同Raft节点的计时器不是同步的。确保`election timeout`不是相同的时间。否则所有的节点都会投票给自己，以至于没有任何一个节点会选举为leader。

当leader可以选举后，我们需要使用Raft保持一致性。我们需要通过`Start()`使服务器接收client操作，并将其插入到log。在Raft，只有leader允许追加log，并通过`AppendEntries`RPC将新的entries传播到其他服务器。

    实现leader和follower追加新log entries。包含`Start()`和`AppendEntries`RPC结构体并发送，填充`AppendEntry` RPC处理程序。你的目标首先应该是通过`TestBasicAgree()`测试用例(`test_test.go`中)。一旦这些工作完成，你应该尝试通过“basic persistence”之前的所有测试用例。
    
* 这个实验的重点是使你的程序可以在各种类型的故障中依然保持健壮性。你需要实现“选举约束”（论文5.4节）。剩下的测试用例会测试各种不同类型的故障，包括不能接收RPC，服务器崩溃及重启。
* Raft的leader是唯一一台能够将entries追加到log中的服务器，其他所有服务器需要独立的给定最新的已提交的entries到他们的本地服务器副本（通过他们自己的`applyCh`）。由于这些，你应该尝试保持着两个行为尽可能的分离。
* 指出当非故障服务器达成一致时Raft使用的消息的最小编号。

基于Raft的服务器必须能够选择何处停止，已经如果计算机重启在何处继续。这要求Raft持久化保存相关状态。(论文中的图2指出了哪些状态需要持久化保存)。

一个真正的实现需要在每次状态发送变化时将Raft的持久化状态写入磁盘。当重启后再从磁盘读取最新保存的数据。你的程序不能使用磁盘，而是从`Persister`(`persister.go`文件)对象中恢复。调用`Make()`提供一个`Persister`对象初始化Raft为最近持久化保存的状态。你可以使用`ReadRaftState()`和`SaveRaftState()`方法读取和保存状态。

    `persist()`序列化所有需要持久化的状态，`readPersist()`反序列化相同的状态。你现在需要决定在何时你的服务器需要持久化这些状态，一旦你的程序完成，你应该通过剩下的测试用例。你可以使用“basic persistence”(`go test -run 'TestPersist1$'`)，然后出来其余的。
    
* 注意：你需要将状态编码为一个byte数组传递给`Persister`,`raft.go`包含了一些关于`persist()`和`readPersist()`示例程序。
* 注意：为了避免内存溢出，Raft必须周期性删除旧的log entries，不过你不需要担心垃圾收集实验中的log。你再下个实验中会通过快照实现。
* 类似于PRC系统只发送结构体中以大写字母开头的属性，忽略小写字母开头的属性。你应该使用GOB编码器持久化保存大写字母开头的属性状态。这是一个隐秘的bug，但是Go不会警告你。
* 为了通过该一些测试用例，你需要优化允许follower北非leader的nextIndex。


    提交之前，请运行所有的测试用例。你应该确保这些程序能够正常工作。记住每次运行可能测试不到一些边角用例，多次运行测试用例是个很好的方法。
```
$ go test
```
    
    
本实验共有17个测试用例：

1. TestInitialElection：启动3个服务器，检查是否只有一个leader；检查未发生故障时term是否会发生变化
2. TestReElection：启动3个服务器，检查是否只有一个leader(L1)；leader(L1)崩溃；检查是否会重新选举一个leader(L2)；leader(L1)恢复，检查leader（L3）；L3和一台服务器崩溃；检查是否没有leader（2台崩溃，超过1/2）;一次恢复崩溃的两台服务器，检查是否有一个leader
3. TestBasicAgree：启动5个服务器，cfg.one写入5个entry，检查是否正确写入cfg的log。（通过<-applyCh写入cfg的log）
4. TestFailAgree：启动3个服务器，写入101；(L1+1)崩溃，写入102,103,104,105；连接崩溃服务器(L1+1),写入106和107，查看崩溃后写入时是否会填充崩溃期间未写入的数据（通过传递prevLog与leader比较，若不同index--，依次循环，直至相同后再一次传递至index；相同则写入）
5. TestFailNoAgree：启动5个服务器，写入10；5台中的3台崩溃（L1+1,L1+2，L1+3）,向L1写入20，检查不会commit；恢复3台服务器（L1+1,L1+2，L1+3），检查leader(L2);向L2写入30，判断写入的位置是否为2；写入1000（index为3），检查是否正确写入（需要填充2）。
6. TestConcurrentStarts：并发写入数据，检查是否正确，需要按照请求顺序依次写入即可
7. TestRejoin：启动3台服务器，写入101；检查leader(L1);L1崩溃，向L1写入102、103、104（都会失败，因为L1故障）；写入103，index应该为2；新leader(L2)故障，L1恢复，写入104；L2恢复，写入105.
8. TestBackup：
9. TestCount
10. TestPersist1
11. TestPersist2
12. TestPersist3
13. TestFigure8
14. TestUnreliableAgree
15. TestFigure8Unreliable
16. TestReliableChurn
17. TestUnreliableChurn
    



