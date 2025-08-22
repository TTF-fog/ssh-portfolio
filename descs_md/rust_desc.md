Rewrite it in Rust!
====
## Why i enjoy it
- There's a couple of nice features i believe should be in other languages, like constant by default and result/err patterns (TO A CERTAIN  DEGREE) 
- the borrowing system is fairly intuitive and acts as a good extension to unique pointers in C++
however, i do have one major gripe with it. When writing in Rust, it genuinely feels like im wrestling
with the compiler. I understand that the Rust is meant to have the compiler act as a overseer where the programmer falls short,
but it at some point it begins to act like an overprotective parent.
For example, take this code snippet
```c++ 
#include <iostream>
int a = 12;
void boo(){
    std::cout << a << std::endl;
}
int main() {
    std::cout << a  << std::endl;;
    a = a + 1;
    boo();
    return 0;
}

```
it declares a global variable, edits it and then accesses it in another function
now, let's look at some equivalent rust code
```rust
static mut A: i32 = 12; //why cant it just be unsafe here

fn boo() {
    unsafe {
        println!("{}", A);
    }
}0

fn main() {
    unsafe {
        println!("{}", A);
        A = A + 1;
        boo();
    }
}
```
yes, i understand that it's not thread-safe - except it's not even meant to be used in threads.
why do i have to mark it as _unsafe_ every **SINGLE** time. oh and there's a bunch of warnings too, despite me explicitly marking it as unsafe.
this is just one example of many.