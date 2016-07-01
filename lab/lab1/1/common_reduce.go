package mapreduce

import (
	"fmt"
	"encoding/json"
	"os"
)
// doReduce does the job of a reduce worker: it reads the intermediate
// key/value pairs (produced by the map phase) for this task, sorts the
// intermediate key/value pairs by key, calls the user-defined reduce function
// (reduceF) for each key, and writes the output to disk.
func doReduce(
	jobName string, // the name of the whole MapReduce job
	reduceTaskNumber int, // which reduce task this is
	nMap int, // the number of map tasks that were run ("M" in the paper)
	reduceF func(key string, values []string) string,
) {
	// TODO:
	// You will need to write this function.
	//你需要使用reduceName(jobName, m, reduceTaskNumber)找到这个reduce任务的中间文件
	//记住你需要对中间文件中的内容进行解码，

	
	//读取文件
	fmt.Printf("mergeFile: %d %d\n",nMap,reduceTaskNumber);

	//根据nMap遍历一下文件，多个map生成针对reduce的文件，
	//读取后解析为key value，保存在map中，key string, values []string
	//最后，针对每个key调用reduceF，获取结果string，即：key-value
	//将最后的key-value写入mergeName的文件中即可
	var maps=make(map[string] []string) //创建一个数组，保存了key-对应关系，给reduceF使用
	//var list=make([]KeyValue,10)
	for i :=0;i<nMap;i++ {
		mergeFile :=reduceName(jobName, i, reduceTaskNumber);
		//读取这个文件内容，再decode下
		// fmt.Printf("mergeFile: %s\n",mergeFile);
		infile, _ := os.Open(mergeFile)	
		dec := json.NewDecoder(infile)

		var v KeyValue
		
		err :=dec.Decode(&v); 
		// fmt.Printf("Decode: %s %s\n",v.Key,err==nil);
		for err==nil{
			//fmt.Printf("Decode: %s %s\n",v.Key,err);

			_,ok :=maps[v.Key]
			if !ok{//未找到
				maps[v.Key]=make([]string,0,1000)
			}
			maps[v.Key]=append(maps[v.Key],v.Value)
			//继续decode
			err =dec.Decode(&v); 
		}

		//解析结果
		defer infile.Close()
	}


	redcueFN :=mergeName(jobName,reduceTaskNumber);
	redcueOF,_ :=os.Create(redcueFN)	
	enc := json.NewEncoder(redcueOF)

	//遍历maps
	for nkey, nvalue := range maps {
		// fmt.Printf("maps: %s %s\n",nkey,nvalue);
		rs :=reduceF(nkey,nvalue)
		// list=append(list,KeyValue{nkey,rs})
		enc.Encode(KeyValue{nkey,rs})
	}
 	redcueOF.Close()

	// You can find the intermediate file for this reduce task from map task number
	// m using reduceName(jobName, m, reduceTaskNumber).
	// Remember that you've encoded the values in the intermediate files, so you
	// will need to decode them. If you chose to use JSON, you can read out
	// multiple decoded values by creating a decoder, and then repeatedly calling
	// .Decode() on it until Decode() returns an error.
	
	//你需要将reduce的输出KeyValue转为JSON并写入文件mergeName()
	//
	// You should write the reduced output in as JSON encoded KeyValue
	// objects to a file named mergeName(jobName, reduceTaskNumber). We require
	// you to use JSON here because that is what the merger than combines the
	// output from all the reduce tasks expects. There is nothing "special" about
	// JSON -- it is just the marshalling format we chose to use. It will look
	// something like this:
	//
	// mergeFile :=mergeName(jobName,reduceTaskNumber);
	// enc := json.NewEncoder(mergeFile)
	// for key in ... {
	// 	enc.Encode(KeyValue{key, reduceF(...)})
	// }
	// file.Close()
}
