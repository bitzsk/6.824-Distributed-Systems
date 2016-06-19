# Go语言学习

GO语言

---

    学习网站：
    http://www.yiibai.com/go/go_start.html
    https://tour.golang.org/welcome/1
    https://tour.go-zh.org/basics/1
    
###输出
fmt.Printf("s[%d] == %s\n", i, s[i])
fmt.Println("%d",2);
fmt.Sprintf("%v (%v years)", p.Name, p.Age) //返回一个字符串
    
###基本类型
Go 的基本类型有Basic types
```
bool
string
int  int8  int16  int32  int64
uint uint8 uint16 uint32 uint64 uintptr
byte // uint8 的别名
rune // int32 的别名 代表一个Unicode码
float32 float64
complex64 complex128
```
int，uint 和 uintptr 类型在32位的系统上一般是32位，而在64位系统上是64位。当你需要使用一个整数类型时，你应该首选 int，仅当有特别的理由才使用定长整数类型或者无符号整数类型。

###零值
变量在定义时没有明确的初始化时会赋值为 零值 。零值是：

* 数值类型为 0 ，
* 布尔类型为 false ，
* 字符串为 "" （空字符串）。


###类型转换
表达式 T(v) 将值 v 转换为类型 T 。
```
var i int = 42
var f float64 = float64(i)//将int转为float64
var u uint = uint(f)//将float64转为uint
//也可以使用如下方式
i := 42
f := float64(i)
u := uint(f)
```

###类型推导
在定义一个变量却并不显式指定其类型时（使用 := 语法或者 var = 表达式语法）， 变量的类型由（等号）右侧的值推导得出。
```
i := 42           // int
f := 3.142        // float64
g := 0.867 + 0.5i // complex128
```

###常量
常量的定义与变量类似，只不过使用 const 关键字

###for
```
for i := 0; i < 10; i++ {
		sum += i
	}
```
不像 C，Java，或者 Javascript 等其他语言，for 语句的三个组成部分 并不需要用括号括起来，但循环体必须用 { } 括起来。

###if
跟 for 一样， if 语句可以在条件之前执行一个简单语句。由这个语句定义的变量的作用域仅在 if 范围之内。
```
func pow(x, n, lim float64) float64 {
	if v := math.Pow(x, n); v < lim {
		return v
	}
	return lim
}
```

###defer
defer 语句会延迟函数的执行直到上层函数返回。延迟调用的参数会立刻生成，但是在上层函数返回前函数都不会被调用。
```
func main() {
	defer fmt.Println("world")
	fmt.Println("hello")
}
```
输出为：
hello
world

**defer栈**
延迟的函数调用被压入一个栈中。当函数返回时， 会按照后进先出的顺序调用被延迟的函数调用。


###指针
Go 具有指针。 指针保存了变量的内存地址。
类型 *T 是指向类型 T 的值的指针。其零值是 nil 。
& 符号会生成一个指向其作用对象的指针。
var p *int =&i
修改指针指向变量的值 *p=23
获取指针指向变量的值 *p 
```
var p *int

i := 42
p = &i //p是指向i的指针

fmt.Println(*p) // 通过指针 p 读取 i
*p = 21         // 通过指针 p 设置 i
```

###结构体
struct就是一个字段的集合,定义方法：
type {Name} struct{...}
初始化 {Name}{1,2}
访问：结构体字段使用点号来访问。
```
type Vertex struct {
	X int
	Y int
}
v := Vertex{1, 2}
v2 = Vertex{X: 1}  // Y:0 被省略
v.X  // is 1
```
**结构体指针**
结构体字段可以通过结构体指针来访问。
通过指针间接的访问是透明的。
```
v:=Vertex{1,2};
p:=&v
p.X=9;
fmt.Println(*p);//is {9 2}
fmt.Println(p); //is &{9 2}
```


###数组
**定义**
类型 [n]T 是一个有 n 个类型为 T 的值的数组。
var a [10]int //定义变量 a
是一个有十个整数的数组。
s := []int{2, 3, 5, 7, 11, 13} //定义一个6个元素的int数组
```
var a [2]string
a[0] = "Hello"
a[1] = "World"
```
**slice**
len(s)可以获取s数组的长度
s[lo:hi]表示从 lo 到 hi-1 的 slice 元素，含前端，不包含后端。
```
s := []int{2, 3, 5, 7, 11, 13}
s[1:4] == [3 5 7]
s[:3] == [2 3 5]
s[4:] == [11 13]
```

**构造 slice**
slice 由函数 make 创建。这会分配一个全是零值的数组并且返回一个 slice 指向这个数组：
a := make([]int, 5)  // len(a)=5
```
a := make([]int, 5)
printSlice("a", a) //a len=5 cap=5 [0 0 0 0 0]
b := make([]int, 0, 5)
printSlice("b", b)//b len=0 cap=5 []
c := b[:2]
printSlice("c", c)//c len=2 cap=5 [0 0]
d := c[2:5]
printSlice("d", d)//d len=3 cap=3 [0 0 0]
```
**向 slice 添加元素**
        
    func append(s []T, vs ...T) []T
append 的第一个参数 s 是一个元素类型为 T 的 slice ，其余类型为 T 的值将会附加到该 slice 的末尾。
```
var a []int
a = append(a, 2, 3, 4,5) //a len=4 cap=4 [2 3 4 5]
```

**range遍历数组**
for 循环的 range 格式可以对 slice 或者 map 进行迭代循环。
```
for i, v := range pow {
	fmt.Printf("2**%d = %d\n", i, v)
}
```
当使用 for 循环遍历一个 slice 时，每次迭代 range 将返回两个值。 第一个是当前下标（序号），第二个是该下标所对应元素的一个拷贝。
range pow返回两个值：i和v
如果不需要可以通过赋值给 _ 来忽略序号和值。
```
for _, value := range pow {
	fmt.Printf("%d\n", value)
}
```
切片更多查看https://blog.go-zh.org/go-slices-usage-and-internals


###map

**初始化**
map在使用之前必须用 make 来创建；值为 nil 的 map 是空的，并且不能对其赋值。

```
//创建一个map，key是字符串，value是结构头
var m map[string]Vertex
m = make(map[string]Vertex)
type Vertex struct {
    Lat, Long float64
}
m["Bell Labs"] = Vertex{
	40.68433, -74.39967,
}
//创建一个map并赋初始值
var m = map[string]Vertex{
	"Bell Labs": Vertex{
		40.68433, -74.39967,
	},
	"Google": Vertex{
		37.42202, -122.08408,
	},
}
//如果顶级的类型只有类型名的话，可以在文法的元素中省略键名。
var m = map[string]Vertex{
	"Bell Labs": {40.68433, -74.39967},
	"Google":    {37.42202, -122.08408},
}
```

**修改map**
```
m[key] = elem   //设置元素
elem = m[key]   //获取元素
delete(m, key)  //删除元素
elem, ok = m[key]   //检测是否存在

delete(m, "Answer")
v, ok := m["Answer"]    //The value: 0 Present? false
```
    
###函数的闭包
Go 函数可以是一个闭包。闭包是一个函数值，它引用了函数体之外的变量。 这个函数可以对这个引用的变量进行访问和赋值；换句话说这个函数被“绑定”在这个变量上。




###方法

```
type Vertex struct {
	X, Y float64
}
//仍然可以在结构体类型上定义方法。
//方法接收者 出现在 func 关键字和方法名之间的参数中。
//(v *Vertex)是方法接受者
//函数定义func Abs() float64 
func (v *Vertex) Abs() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y)
}

func main() {
	v := &Vertex{3, 4}
	fmt.Println(v.Abs())//5
}
```

你可以对包中的 任意 类型定义任意方法，而不仅仅是针对结构体。
```
type Integer int32
func (i Integer) Abs() int32{
  if(i<0){
  	return int32(-i)
  }
  return int32(i)
}
ik := Integer(-3)
fmt.Println(ik.Abs())
```


###接口
```
//定义一个Abser的接口，内部有方法Abs() float64
type Abser interface {
	Abs() float64
}
```

```
type Abser interface {
	Abs() float64
}

func main() {
	var a Abser
	f := MyFloat(-math.Sqrt2)
	v := Vertex{3, 4}

	a = f  // a MyFloat 实现了 Abser
	a = &v // a *Vertex 实现了 Abser

	// 下面一行，v 是一个 Vertex（而不是 *Vertex）
	// 所以没有实现 Abser。
	//a = v

	fmt.Println(a.Abs())
}

type MyFloat float64
//为MyFloat 实现Abs()
func (f MyFloat) Abs() float64 {
	if f < 0 {
		return float64(-f)
	}
	return float64(f)
}
//定义结构体
type Vertex struct {
	X, Y float64
}
//为*Vertex 实现Abs()
func (v *Vertex) Abs() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y)
}
```


**隐式接口**
类型通过实现那些方法来实现接口。 没有显式声明的必要；所以也就没有关键字“implements“。隐式接口解藕了实现接口的包和定义接口的包：互不依赖。

###Stringers
一个普遍存在的接口是 fmt 包中定义的 Stringer。类似java对象中的toString()
```
//Stringer 是一个可以用字符串描述自己的类型。`fmt`包 （还有许多其他包）使用这个来进行输出。
type Stringer interface {
    String() string
}
//例如，a和z是结构体Person的实例，
a := Person{"Arthur Dent", 42}
z := Person{"Zaphod Beeblebrox", 9001}
fmt.Println(a, z)//打印时会输出Person的String()

func (p Person) String() string{
}
```

##错误
Go 程序使用 error 值来表示错误状态。与 fmt.Stringer 类似， error 类型是一个内建接口：
```
type error interface {
    Error() string
}
```
（与 fmt.Stringer 类似，fmt 包在输出时也会试图匹配 error。）
通常函数会返回一个 error 值，调用的它的代码应当判断这个错误是否等于 nil， 来进行错误处理。
```
i, err := strconv.Atoi("42")
//error 为 nil 时表示成功；非 nil 的 error 表示错误。
if err != nil {
    fmt.Printf("couldn't convert number: %v\n", err)
    return
}
fmt.Println("Converted integer:", i)
```



###Readers
io 包指定了 io.Reader 接口， 它表示从数据流结尾读取。
Go 标准库包含了这个接口的许多实现， 包括文件、网络连接、压缩、加密等等。

io.Reader 接口有一个 Read 方法：
```
func (T) Read(b []byte) (n int, err error)
```
Read 用数据填充指定的字节 slice，并且返回填充的字节数和错误信息。 在遇到数据流结尾时，返回 io.EOF 错误。


###Web 服务器
包 http 通过任何实现了 http.Handler 的值来响应 HTTP 请求：
```
package http

type Handler interface {
    ServeHTTP(w ResponseWriter, r *Request)
}
```