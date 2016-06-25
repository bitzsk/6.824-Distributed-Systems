package mapreduce

import (
	"hash/fnv"
	"os"
	// "io"
	// "bufio"
	"encoding/json"
	"fmt"
)

// doMap does the job of a map worker: it reads one of the input files
// (inFile), calls the user-defined map function (mapF) for that file's
// contents, and partitions the output into nReduce intermediate files.
func doMap(
	jobName string, // the name of the MapReduce job
	mapTaskNumber int, // which map task this is
	inFile string,
	nReduce int, // the number of reduce task that will be run ("R" in the paper)
	mapF func(file string, contents string) []KeyValue,
) {
	//hehe:test 0 824-mrinput-0.txt 1
	fmt.Printf("doMap:%s %d %s %d\n",jobName,mapTaskNumber,inFile,nReduce);
	file, err := os.Open(inFile)
	if err != nil {
		fmt.Printf("open file " + inFile + " failed!")
		return
	}
	defer file.Close()
	
	// br := bufio.NewReader(file)
	// ret := string("")
	
	// for {

	// 	line, _, err1 := br.ReadLine() //读取一行
	// 	if err1 !=nil{
	// 		if err1!=io.EOF{
	// 			break
	// 		}
	// 	}
	// 	if len(line)>999{
	// 		break;
	// 	}
	// 	str := string(line)  //将byte数组转为字符串
	// 	ret =ret + str
	// 	fmt.Printf("hehe:%s\n",str);
	// }
	ret :=string("")
	buf := make([]byte, 1024)
    for{
            n, _ := file.Read(buf)
            if 0 == n { break }
            // os.Stdout.Write(buf[:n])
            ret+=string(buf[:n]);
    }
	
	// fmt.Printf("%s\n",ret);
	kv :=mapF(inFile,ret)
	fmt.Printf("%s %d\n","after mapF ",len(kv));

	var kValues=make([][]KeyValue,nReduce);

	for _,val :=range kv{
		uv :=ihash(val.Key);
		uv=uv%uint32(nReduce)
		
		//根据hash获取了不同所属的文件名
		kValues[uv]=append(kValues[uv],val)
	}

	//目前，已经读取了文件及其内容，
	//处理后为KeyValue数组,结构体，定义在common.go中
	//需要生成nReduce个文件，命名规则：jobName+index

	//mapF(inFile,);
	// TODO:
	// You will need to write this function.
	// You can find the filename for this map task's input to reduce task number
	// r using reduceName(jobName, mapTaskNumber, r). The ihash function (given
	// below doMap) should be used to decide which file a given key belongs into.

	//inFile是读取的文件，
	//ihash用来决定指定文件的key

	//考虑一下如何对中间文件命名，需要明确看出map任务和reduce任务
	// The intermediate output of a map task is stored in the file
	// system as multiple files whose name indicates which map task produced
	// them, as well as which reduce task they are for. Coming up with a
	// scheme for how to store the key/value pairs on disk can be tricky,
	// especially when taking into account that both keys and values could
	// contain newlines, quotes, and any other character you can think of.
	
	// 作为reduce任务的输出 必须是json格式
	// One format often used for serializing data to a byte stream that the
	// other end can correctly reconstruct is JSON. You are not required to
	// use JSON, but as the output of the reduce tasks *must* be JSON,
	// familiarizing yourself with it here may prove useful. You can write
	// out a data structure as a JSON string to a file using the commented
	// code below. The corresponding decoding functions can be found in
	// common_reduce.go.
	//
	  // enc := json.NewEncoder(file)
	  // for _, kv := ... {
	  //   err := enc.Encode(&kv)
	//
	// Remember to close the file after you have written all the values!
	//记得最后关闭文件
	

	for i :=0;i<nReduce ;i++{
		
		fname :=reduceName(jobName,mapTaskNumber,i)
		fmt.Printf("fname: %s\n",fname)
		outfile ,_:= os.Create(fname);

		defer outfile.Close();
		enc := json.NewEncoder(outfile)
		  for _, k_v := range kValues[i] {
		    _ = enc.Encode(&k_v)
		}
	}
	
}

func ihash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}
