# locks
Rlock will wait for lock before running, but many Rlocks could occur at the same time

Lock will wait for all Rlocks to clear before proceeding

use lock for writing and Rlock for reading


# concurency
each go will take up a core, there can only be 8 go concurent threads working if i have a 8 core CPU

# new vs make
*T: address to a type, e.g. *int address to a int variable

p := new(T)
What it returns:
Type: *T
Value: pointer to zero-value T

e.g. 
type User struct {
    age int
}

u := new(User)
fmt.Println(u)       // &{0}
fmt.Println(u.age)   // 0

make is only used for slice, map, and channel

when you create a map variable, it is nil and you can't use it. So we would use make() to create the object

var m map[string]int
m["a"] = 1   // ❌ panic

m := make(map[string]int)
m["a"] = 1   // ✅ works

# channel
defer close(c) - defer keyword runs the code right before function ends

x := <- ch is blocked and waits for ch <- 42, then they run at the same time

in buffered ch, since many values could occur, ch <- 42 could run first, then x := <- 42 could occur later, but if there's nothing in ch, x := <- ch will still wait

when using:
for i := range ch{
    print(i)
}
this wouldn't work because for loop will keep listening for ch even if it is empty. Thus we use close(ch) to tell ch to not receive anything anymore but ch could still be read from. Then the for loop could exit

select{
    case c := <- chickenChannel:
        ...
    case b := <- beefChannel:
        ...
}
if chicken or beef channel receives anything, the corresponding block of code will run