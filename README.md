# 课程介绍

---
    6.824是一个由12个单元组成的科目，包括讲课、阅读、编程实验、一个可选的项目、期中和期末考试组成。

**什么是分布式系统？**

* 多个协同工作的计算机
* DNS服务器、P2P文件共享系统、大型数据库、MapReduce模型等
* 许多关键的IT基础设施都是分布式系统

**如何分布式**

* 连接物理上隔离的计算机
* 通过隔离获得安全性
* 通过复制容忍错误
* 通过并行CPU/内存/磁盘/网络 提高生产力

**缺点**

* 由于需要做许多并行处理，增加了复杂性
* 必须处理局部的错误
* 难以充分利用潜在的性能

**为什么学习这堂课程**

* 对困难问题和无明显解决方案的问题感兴趣
* 需要在实际系统中应用分布式，比如：网站的增长导致需要分布式系统
* 活跃的研究领域：分布式领域有很多未解决的问题，而且还在不断增加
* 动手实践：你将会自动动手完成一个分布式系统

[http://pdos.csail.mit.edu/6.824](http://pdos.csail.mit.edu/6.824)

课程包含交流、论文和实验

阅读：在课程开始前需要先阅读论文，每个论文有一个简短的问题。

实验目标：深入理解一些重要的技术；分布式编程经验

* 实验 1: MapReduce
* 实验 2: 通过复制容错
* 实验 3: 容错的key/value存储
* 实验 4: 共享的key/value存储

实验完成的标准：测试用例通过

##主要内容

这是一个关于IT基础设施的课程，这些基础设施主要为应用提供服务。
    课程的概述部分跳过了分布式系统中比较复杂的细节，主要包括重要的三个部分：
    
* 存储
* 通信
* 计算

**1.实现**
  分布式系统的实现所需的技术包括RPC(远程程序调用)、threads（线程）、并发控制

**2.性能**
性能的目标是：可伸缩的计算能力
Nx servers -> Nx total throughput via parallel CPU, disk, net.
分布式系统强大的处理能力是通过并行CPU,磁盘和网络达成的。所以，更多的负载只能依靠购买更多的计算机。
分布式系统的伸缩性变得越来越难，因为：

    * 负载均衡、离群者。
    * 一些较小的非并行的部分
    * 隐蔽的共享资源，如网络

**3.错误容忍**
在成百上千的服务器、复杂的网络系统中，通常总会有一些不能正常工作的资源。分布式系统在为应用提供服务时需要隐藏这些错误，我们通常关注的是：

    * 可用性  随时可以使用文件而不管是否发生错误
    * 持久性  错误被修复后可以正常工作

解决方法：将数据复制到其他服务器，如果一个服务器崩溃后，客户端可以使用其他服务器。

**4.一致性**
通用性的基础设施需要非常良好的定义其行为。例如：Get(k)从最新的Put(k,v)获取值。
但是，实现良好的行为是很难的！因为：

    * 客户端提交是并行的操作
    * 服务器会在某些时刻崩溃
    * 网络可能不能连接到正常运行的服务器

除此之外，一致性和性能也很难达到：

    * 一致性需要通信，例如：获得最新的Put()值
    * 令人满意的系统（严格的系统）往往是缓慢的。
    * 快速的系统通常使得应用需要复杂的行为。

##案例：MapReduce
https://pdos.csail.mit.edu/6.824/papers/mapreduce.pdf
http://wenku.baidu.com/view/06b48cedaeaad1f346933f6e.html?from=search

    MR是6.824分布式课程的典型范例，实验一中我们将详细了解MR的使用。
    
**MapReduce综述**

* 使用场景：使用多台计算机处理TB级别的数据集，例如：在分析网络爬虫抓取的海量网站页面时，如果没有“好心人”开发的分布式系统，会让人很痛苦（比如：复制时发生的错误）。
* 主要目标：方便非专业的开发者可以有效的分析处理大量数据。在MR中，开发者只需要定义Map和Reduce函数，这相当容易。MR模型框架会将这些代码运行在数以千计台机器上，因此整个系统输入量巨大，MR框架同时还隐藏了分布式的所有细节，使得开发者只需关注业务逻辑的处理。

**MapReduce概览**
输入被拆分为“splits” 
```ruby
Input Map -> a,1 b,1 c,1
Input Map ->     b,1
Input Map -> a,1     c,1
              |   |   |
                  |   -> Reduce -> c,2
                  -----> Reduce -> b,2
```
MR调用在每个split（分割后的数据集）上调用Map()，产生了集合(k2,v2)的中间数据,如（a,1）。
MR收集所有key为k2的中间数据，将其值v2传递个Reduce
最后，Reduce输出了<k2,v3>对。如，Reduce收集所有key为c的中间数据，求和后得出(c,2)

例如：字数统计,输入是许多文本文件
```javascript
Map(k, v)
    split v into words
    for each word w
      emit(w, "1")
 Reduce(k, v)
    emit(len(v))
```
MapReduce模型很容易编程，他隐藏了许多令人痛苦的细节：
* 并发--与顺序执行具有同样的结果
* 开始服务器的s/w
* 数据传输
* 错误处理

MapReduce模型的伸缩性很好。
Map()函数并不会相互等待或共享数据，可以并行处理。Reduce()函数也是一样。所以，更多的机器可以提供更多的处理能力。

**性能的限制因素**
系统优化时，我们所关心的是有CPU、内存、磁盘和网络？
对于分布式系统来说，性能主要受制于Map服务器和reduce服务器之间用于通信的网络系统所具有的“带宽”。网络系统总的吞吐量要小于所有服务器网络连接速度的总和。
很难构建一个具有150MB/s带宽的网络系统。所以我们需要关注的是减少数据在网络上的传输。

**错误容忍**
服务器在运行MR任务时崩溃时会发生什么？隐藏错误是一项巨大的工程。为什么不重新开始重启整个任务？
MR只在map()和reduce()失败时重新运行，因为他们不会修改输入，也不会保持状态，没有共享内存，也没有map和map，reduce和reduce之间的交互。因此重新执行时生成的最终结果不会发生变化。
相对于其他的并行编程模型，MR很简单，只需要给定相对独立的两类函数。

**更多细节**

* master：向workers分配任务；记住splits在map处理后生成的中间输出。
* MR处理的输入一般存储在GFS中，每个split有3份拷贝。
* 所有的计算机同时运行GFS和MR workers。
* 输入数据splits要比workers多
* master启动每台服务器上的Map任务，并在任务执行结束后为其分发新的任务。
* worker将Map的输出hash到R个部分，并放在本地硬盘上。
* 只有当所有Map任务都结束后才能开始Reduce任务。
* master通知执行Reduce的服务器从Map worker中拉取中间数据片。
* Reduce worker将最终结果写入GFS

**如何充分利用网络** 

* Map的输入从本地磁盘中读取，而不是通过网络传输获取。
* 中间数据只通过网络传输一次，存储在本地磁盘而非GFS。
* 中间数据分割为多个文件，大量的网络传输效率更高。

**如何负载均衡**

* 伸缩性的关键是NX servers 初次之外别无他法
* 处理split和数据片的时间不一致，这是因为不同的服务器处理的大小和内容不同，硬件也不尽相同。
* 解决方案：split的数量比worker多。Master会将最新的split分发送给完成任务的worker；split都很小，不会从开始一直运行到结束；因此速度快的服务器相比慢速服务器来说，运行了更多的任务，综合来看，花费了相同的时间。

**MR复制时worker崩溃**

* Map worker崩溃：
    * master重新运行，读取GFS上输入数据的副本，即使worker执行完成，或者已经将中间数据写入本地磁盘。
    * Reduce workers 可能已经读入了失败后的中间数据
    * 我们如何知道worker崩溃（依靠ping吗）
    * 如果worker在Reduces拉取所有中间数据后崩溃，则不需要重新运行，否则Reduce失败后需要等待Map重新执行
* Reduce worker在生成输出之前崩溃，master需要在其他worker上重新启动任务。
* Reduce worker 在写入输出期间崩溃时的处理：GFS会保持写入的原子性，在写入完成之前，GFS会隐藏这些内容，写入完成后才会重命名。所以，master重新运行Reduce任务时安全的。

**其他问题**

* master意外地启动了两个Map worker对于同一份输入进行处理时该如何处理？
    * MR只会告诉Reduce workers其中一个的输出。
*两个Reduce worker处理同一份中间数据片时如何处理？
    * MR将会试着向GFS写入相同的输出文件，由于GFS具有原子性，第二次写入会生效。
* 某个worker运行的很慢时如何处理？
   * 这可能是硬件发生故障了，master启动最少任务的第二份拷贝。
* 某个worker计算了错误的输出时如何处理？
    * 太糟糕了，MR假定了CUP和软件不能停止
* 如果master崩溃了怎么办？

**MapReduce不适用于的应用**
并非所有的事情都适合map/shuffle/reduce模式。
处理少量数据，因为管理调度的会很花费时间。例如：网站的后台不需要使用MR
将小的更新加入到大数据中。例如：将一个小文档添加到大索引中
不可预知的输入（Map和Reduce都不能选择输入）
重复的shuffles数据。例如：页面rank（虽然可以使用多个MR，但效率不高）


**结论**
MapReduce使得大量集群系统变得流行起来
*  并非最有效和最灵活
*  良好的伸缩性
*  易于编程 --隐藏了错误和数据传输

在实践中，有很多需要权衡的地方。





https://pdos.csail.mit.edu/6.824/notes/l01.txt

