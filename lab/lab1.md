# 实验一：MapReduce

---

###介绍
在本节实验中，我们将通过创建一个MapReduce库来学习Go语言，并学习分布式系统如何进行错误处理。第一部分中，需要编写一个简单的MapReduce程序。第二部分，需要写一个Master为workers分发任务并处理workers崩溃的程序。错误处理的方法与MapReduce论文中的方式十分相似。

###软件
我们将使用Go语言实现这些库。实验提供的MapReduce程序既包括分布式的，也包括非分布式的。你可以通过git获取原始的库文件，（git是一种版本控制程序，可以通过http://www.kernel.org/pub/software/scm/git/docs/user-manual.html学习如何使用git）。如果你已经对版本控制系统有过了解，可以通过http://eagain.net/articles/git-for-computer-scientists/深入了解其原理。

本课程git仓库地址是
        
    git://g.csail.mit.edu/6.824-golabs-2016

**安装**
通过下面的命令安装
```
$ add git # only needed on Athena machines
$ git clone git://g.csail.mit.edu/6.824-golabs-2016 6.824
$ cd 6.824
$ ls
Makefile src
```
Git可以帮助你跟踪代码的修改，例如，如果你想要将当前进度保存为可恢复的检查点，可以通过下面的命令提交修改
```
$ git commit -am 'partial solution to lab 1'
```
实验提供的Map/Reduce实现支持两种模式的操作，顺序式和分布式。过去，map和reduce任务都是串行执行：

* 首先，第一个map任务执行完毕，再执行第二个，之后执行第三个....
* 当所有的map任务都完成后，第一个reduce任务开始运行，然后第二个reduce任务，一次类推。
* 这个模型的运行速度并不是很快，但是很容易调试，因为没有并行执行带来的噪音。

对于分布式：
* 分布式模型运行了很多worker线程，首先并行执行map任务，再并行执行reduce任务。这个模型虽然很快，但是实现起来会更加复杂，也难于调试。
        
    

###前言：熟悉源程序
mapreduce包提供了一个简单的顺序式Map/Reduce库。应用程序一般应该调用Distributed() [master.go] 启动分布式执行的任务，也可以调用Sequential()[master.go]进行顺序执行，这样可便于调试。

mapreduce实现流程

1. 应用提供多个输入文件，一个map函数，一个reduce函数，以及多个reduce任务(nReduce)。
2. master 启动RPC服务[master_rpc.go],等待 workers注册（使用RPC调用Register()[master.go]）。当任务可用时(步骤4和5),schedule() [schedule.go]决定如何分配这些任务到相应的worker中，并决定如何处理worker崩溃。
3. master通过指定map任务处理每个输入文件，每个任务至少执行一次doMap()[common_map.go] 。master直接执行（当使用Sequential()）或者将DoTask通过RPC分配给worker [worker.go]。每次调用doMap()读取适合的文件，调用map函数在处理这些文件，分别产生nReduce个文件。因此，在所有的map任务执行完毕后，将会有files x nReduce个文件
    
        f0-0, ..., f0-0, f0-, ...,
        f<#files-1>-0, ... f<#files-1>-.
4.  master针对每个reduce任务调用至少一次doReduce() [common_reduce.go]。类似doMap()，可以自己执行或分配给worker执行。doReduce() 会从每个map(f-*-)收集nReduce个reduce文件，之后在这些文件上调用reduce函数，将会产生nReduce个结果文件。
5.  master调用mr.merge() [master_splitmerge.go]，合并所有步骤4产生的nReduce个结果文件。
6.  master通过RPC向每个worker调用关闭程序，然后关闭自己的RPC服务。
        
        注意：本课程的练习中，你将只需要修改doMap, doReduce, 和 schedule。这几个方法分别对应着common_map.go, common_reduce.go, 和 schedule.go。你还需要在../main/wc.go文件中编写自己的map和reduce方法。

虽然你在做练习时并不需要修改其他文件，但是阅读这些文件可对于理解这些方法和全局架构很有帮助。

###任务一: Map/Reduce输入和输出
实验给出的Map/Reduce实现中缺少了某些部分。在你可以写出自己的第一个Map/Reduce方法对之前，你需要进行一些修改以完成Map/Reduce的顺序实现。
需要注意的是，我们提供的代码缺失了两个关键部分：用于划分map任务输出文件的方法和为reduce任务收集所有输入文件的方法。
这些任务通过common_map.go文件中的doMap()以及common_reduce.go中的doReduce()方法进行。这些文件的注释会为你指明方向的。
为了帮助你判断是否正确实现了doMap() 和 doReduce()，我们提供了Go语言的测试套件来检查你的程序。这些测试用例在test_test.go。启动这些检查代码是否正确的用例，可以执行如下命令:
```
$ cd 6.824
$ export "GOPATH=$PWD"  # go needs $GOPATH to be set to the project's working directory
$ cd "$GOPATH/src/mapreduce"
$ setup ggo_v1.5
$ go test -run Sequential mapreduce/... 
ok      mapreduce   2.694s
```

    如果你的程序通过了所有的Sequential用例，可以获得满分。
如果你的输出不是OK，这就意味着你的程序有bug。在common.go中设置debugEnabled = true，并在命令行中添加 -v
```
go test -v -run Sequential mapreduce/... 
如果不信的话执行go test -v -run Sequential mapreduce/..
若内存溢出，可以修改test_test.go中的nNumber为10000
```


###任务二:单worker统计单词
现在，map和reduce已经被修复了。我们可以实现一些有意思的Map/Reduce操作。这部分我们将实现一个统计单词文字的例子，你的任务时修改mapF 和reduceF，让wc.go报告每个单词出现的数量。判断是否是单词可以通过[unicode.IsLetter](http://golang.org/pkg/unicode/#IsLetter)。
`~/6.824/src/main`下有很多`pg-*.txt`的文件,编译和启动软件的方法：
```
$ cd 6.824
$ export "GOPATH=$PWD"
$ cd "$GOPATH/src/main"
$ go run wc.go master sequential pg-*.txt
# command-line-arguments
./wc.go:14: missing return at end of function
./wc.go:21: missing return at end of function
```
上面的编译失败是因为wc.go中的mapF()和reduceF()没有编写完成。在编写代码之前，请先阅读[MapReduce paper的Section 2](http://research.google.com/archive/mapreduce-osdi04.pdf)。
你的`mapF()`和`reduceF()`函数与论文中的Section 2.1略有不同。你的`mapF()`会传入相同的文件名和文件内容；应该先将这些字符串拆分为单词，返回key/value对（mapreduce.KeyValue）。你的`reduceF()`针对每个key会被调用一次，传递的参数是所有的mapF()返回的值。
你可以测试的程序是否正确：
```
$ cd "$GOPATH/src/main"
$ time go run wc.go master sequential pg-*.txt
master: Starting Map/Reduce task wcseq
Merge: read mrtmp.wcseq-res-0
Merge: read mrtmp.wcseq-res-1
Merge: read mrtmp.wcseq-res-2
master: Map/Reduce task completed
14.59user 3.78system 0:14.81elapsed
```
输出文件是mrtmp.wcseq，你可以通过下面的命令删除这个文件：
```
$ rm mrtmp.*
```
如果你的实现正确，会出现如下内容：
```
$ sort -n -k2 mrtmp.wcseq | tail -10
he: 34077
was: 37044
that: 37495
I: 44502
in: 46092
a: 60558
to: 74357
of: 79727
and: 93990
the: 154024
```
为了测试方便，可以运行：
```
$ sh ./test-wc.sh
```
该脚本将报告你的程序是否正确。

    * 提示：(Go Blog on strings)[http://blog.golang.org/strings]可以帮助理解Go语言的字符串
    * 你可以使用strings.FieldsFunc拆分字符串
    * [strconv](http://golang.org/pkg/strconv/)包可以将字符串转为整型。

###任务三:分布式MapReduce任务
Map/Reduce的一个卖点就是：开发者无需关注他们的程序是如何并行地运行在多台服务器上面。理论上，任务二的程序会自动并行执行。

我们当前的只实现串行运行所有的map和reduce任务，虽然概念上简单，但是运行效率并不佳。任务三中，你需要完成一个新的版本，这个版本中，MapReduce 为了充分利用多核，会将任务分发到多个worker线程中。由于我们不能模拟真实的Map/Reduce部署环境，只能采用RPC和channels模拟分布式计算机。

为了协调任务的并行执行，我们使用一个特殊的master线程，该线程将任务分发到workers并等待其处理完成。为了使实验更加真实，master应该只能通过RPC与workers进行交互。实验已经提供了worker程序（mapreduce/worker.go），改程序会启动workers并处理RPC消息（mapreduce/common_rpc.go）。

你的任务是完成mapreduce包中的schedule.go，特别是你要修改schedule.go文件中的schedule()，将map和reduce任务分发给workers，并在所有任务都结束后才返回。

master.go中的`run()`会调用你的`schedule()`运行map和reduce任务，然后，调用`merge()`将每一个reduce文件的输出合并到一个单独的输出文件中。`schedule`只需要告诉workers初始的输入文件(`mr.files[task]`)和任务。每个worker都知道从哪个文件中读取自己的输入，也清楚处理完成后写入到哪个输出文件中。master通过RPC调用通知worker新的任务（Worker.DoTask），参数为DoTaskArgs 。

当worker启动后，需要向master发送Register RPC请求。`mapreduce.go`已经实现了master的`Master.Register`的处理程序，传递新worker的信息给`mr.registerChannel`。你的`schedule`应该处理新的worker注册。

`master.go`中的`Master`结构体保存了当前运行的任务信息。注意：master不需要Map和Reduce的哪个函数在使用，workers会注意执行正确的代码。

为了测试你的程序，你应该使用任务一中的测试套件，将-run Sequential修改为-run TestBasic.
```
$ go test -run TestBasic mapreduce/...
```

    * master应该并行的发送RPC给workers，这样workers才能并行处理任务。Go语言的RPC包（http://golang.org/pkg/net/rpc/）
    * master在所有任务都分发出去后，可能需要等待worker任务执行完成。channels对应同步多线程很有用，Channels可以查看（http://golang.org/doc/effective_go.html#concurrency）
    * 并行时最简单的调试方法是在程序中插入debug()语句，并执行`go test -run TestBasic mapreduce/... > out`将输出打印至out文件中。

###任务四：处理worker异常
本任务中，你需要帮助master对发生故障的worker进行处理。MapReduce的异常处理相对简单，这是因为worker不需要保存状态。如果worker节点崩溃，任何从master到worker的RPC都会失败。master应该将失败的任务发给其他worker。
一次RPC失败并不意味着worke挂掉了；worker可能只是暂时未能通信，但仍在计算。因此，可能有这种状况发生：两个worker接受到了相同的任务并计算之。不管怎样，由于任务时幂等的（任意多次执行所产生的影响均与一次执行的影响相同），一个任务计算两次的情况不需要进行处理。（我们的测试用例不会在任务处理过程中失败，所以你不用担心多个worker写入相同的输出文件）

    注意：我们假设master不会发生故障，所以你不需要处理master的故障。master的错误容忍机制要比worker更困难，因为master保存了很多状态，如果发生故障，需要顺序恢复。很多实验室都致力于这项挑战。
    
你的程序必须通过test_test.go中两个剩下的测试用例。第一个用例测试一个worker发生故障；第二个用例测试多个worker发生故障时的处理。测试用例会定期启动新worker，master为其分配任务，但是这些worker处理任务一会就会发生故障。输入下列命令可以运行用例：
```
$ go test -run Failure mapreduce/...
```

###任务五：生成反向索引（可选）

统计单词是一个经典的Map/Reduce应用，但是使用这个Map/Reduce应用的用户不多。我们很少会在一个真实的大数据集合中统计单词。这个任务中，我们将会创建Map和Reduce函数，用来生成反向索引。

反向索引在计算机科学中应用十分广泛，对于文档搜索来说特别有用。宽泛的说，反向索引是一个关于基础数据的map，指向数据初始的位置。例如：在上下文搜索中，map保存的内容是keyword与包含这些单词的文档的映射。

我们已经创建了`main/ii.go`，非常类似之前创建的`wc.go`。你需要修改`main/ii.go`中的`mapF`和`reduceF`，用来生成反向索引。运行`ii.go`应该会输出元数组，每行一个，格式如下：

```
$ go run ii.go master sequential pg-*.txt
$ head -n5 mrtmp.iiseq
A: 16 pg-being_ernest.txt,pg-dorian_gray.txt,pg-dracula.txt,pg-emma.txt,pg-frankenstein.txt,pg-great_expectations.txt,pg-grimm.txt,pg-huckleberry_finn.txt,pg-les_miserables.txt,pg-metamorphosis.txt,pg-moby_dick.txt,pg-sherlock_holmes.txt,pg-tale_of_two_cities.txt,pg-tom_sawyer.txt,pg-ulysses.txt,pg-war_and_peace.txt
ABC: 2 pg-les_miserables.txt,pg-war_and_peace.txt
ABOUT: 2 pg-moby_dick.txt,pg-tom_sawyer.txt
ABRAHAM: 1 pg-dracula.txt
ABSOLUTE: 1 pg-les_miserables.txt
```
如果还不清楚上面的list，格式是：

```
word: #documents documents,sorted,and,separated,by,commas
```
要想通过这项任务，你需要运行`test-ii.sh`:

```
$ sort -k1,1 mrtmp.iiseq | sort -snk2,2 mrtmp.iiseq | grep -v '16' | tail -10
women: 15 pg-being_ernest.txt,pg-dorian_gray.txt,pg-dracula.txt,pg-emma.txt,pg-frankenstein.txt,pg-great_expectations.txt,pg-huckleberry_finn.txt,pg-les_miserables.txt,pg-metamorphosis.txt,pg-moby_dick.txt,pg-sherlock_holmes.txt,pg-tale_of_two_cities.txt,pg-tom_sawyer.txt,pg-ulysses.txt,pg-war_and_peace.txt
won: 15 pg-being_ernest.txt,pg-dorian_gray.txt,pg-dracula.txt,pg-frankenstein.txt,pg-great_expectations.txt,pg-grimm.txt,pg-huckleberry_finn.txt,pg-les_miserables.txt,pg-metamorphosis.txt,pg-moby_dick.txt,pg-sherlock_holmes.txt,pg-tale_of_two_cities.txt,pg-tom_sawyer.txt,pg-ulysses.txt,pg-war_and_peace.txt
wonderful: 15 pg-being_ernest.txt,pg-dorian_gray.txt,pg-dracula.txt,pg-emma.txt,pg-frankenstein.txt,pg-great_expectations.txt,pg-grimm.txt,pg-huckleberry_finn.txt,pg-les_miserables.txt,pg-moby_dick.txt,pg-sherlock_holmes.txt,pg-tale_of_two_cities.txt,pg-tom_sawyer.txt,pg-ulysses.txt,pg-war_and_peace.txt
words: 15 pg-dorian_gray.txt,pg-dracula.txt,pg-emma.txt,pg-frankenstein.txt,pg-great_expectations.txt,pg-grimm.txt,pg-huckleberry_finn.txt,pg-les_miserables.txt,pg-metamorphosis.txt,pg-moby_dick.txt,pg-sherlock_holmes.txt,pg-tale_of_two_cities.txt,pg-tom_sawyer.txt,pg-ulysses.txt,pg-war_and_peace.txt
worked: 15 pg-dorian_gray.txt,pg-dracula.txt,pg-emma.txt,pg-frankenstein.txt,pg-great_expectations.txt,pg-grimm.txt,pg-huckleberry_finn.txt,pg-les_miserables.txt,pg-metamorphosis.txt,pg-moby_dick.txt,pg-sherlock_holmes.txt,pg-tale_of_two_cities.txt,pg-tom_sawyer.txt,pg-ulysses.txt,pg-war_and_peace.txt
worse: 15 pg-being_ernest.txt,pg-dorian_gray.txt,pg-dracula.txt,pg-emma.txt,pg-frankenstein.txt,pg-great_expectations.txt,pg-grimm.txt,pg-huckleberry_finn.txt,pg-les_miserables.txt,pg-moby_dick.txt,pg-sherlock_holmes.txt,pg-tale_of_two_cities.txt,pg-tom_sawyer.txt,pg-ulysses.txt,pg-war_and_peace.txt
wounded: 15 pg-being_ernest.txt,pg-dorian_gray.txt,pg-dracula.txt,pg-emma.txt,pg-frankenstein.txt,pg-great_expectations.txt,pg-grimm.txt,pg-huckleberry_finn.txt,pg-les_miserables.txt,pg-moby_dick.txt,pg-sherlock_holmes.txt,pg-tale_of_two_cities.txt,pg-tom_sawyer.txt,pg-ulysses.txt,pg-war_and_peace.txt
yes: 15 pg-being_ernest.txt,pg-dorian_gray.txt,pg-dracula.txt,pg-emma.txt,pg-great_expectations.txt,pg-grimm.txt,pg-huckleberry_finn.txt,pg-les_miserables.txt,pg-metamorphosis.txt,pg-moby_dick.txt,pg-sherlock_holmes.txt,pg-tale_of_two_cities.txt,pg-tom_sawyer.txt,pg-ulysses.txt,pg-war_and_peace.txt
younger: 15 pg-being_ernest.txt,pg-dorian_gray.txt,pg-dracula.txt,pg-emma.txt,pg-frankenstein.txt,pg-great_expectations.txt,pg-grimm.txt,pg-huckleberry_finn.txt,pg-les_miserables.txt,pg-moby_dick.txt,pg-sherlock_holmes.txt,pg-tale_of_two_cities.txt,pg-tom_sawyer.txt,pg-ulysses.txt,pg-war_and_peace.txt
yours: 15 pg-being_ernest.txt,pg-dorian_gray.txt,pg-dracula.txt,pg-emma.txt,pg-frankenstein.txt,pg-great_expectations.txt,pg-grimm.txt,pg-huckleberry_finn.txt,pg-les_miserables.txt,pg-moby_dick.txt,pg-sherlock_holmes.txt,pg-tale_of_two_cities.txt,pg-tom_sawyer.txt,pg-ulysses.txt,pg-war_and_peace.txt
```

###运行所有用例

你可以通过运行脚本`src/main/test-mr.sh`执行所有测试用例。若程序正确，输出如下：
```
$ sh ./test-mr.sh
==> Part I
ok      mapreduce   3.053s

==> Part II
Passed test

==> Part III
ok      mapreduce   1.851s

==> Part IV
ok      mapreduce   10.650s

==> Part V (challenge)
Passed test
```




**注意事项**

若需要在在windows下运行，需要进行如下修改：
    
    *   由于不支持unix domain，可以使用tcp。修改test_test.go中port即测试用例使用的相关代码，返回"localhost:8888"类似的地址
    *   将net.Listen("unix", me)修改为修改net.Listen("tcp", me)
    *   可以对const nNumber  10000nMap nReduce 进行修改，加快测试速度。

