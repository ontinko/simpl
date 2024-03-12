# Simpl

Small programming language made to learn how a programming language works, **WIP**

## Code example

```
# variable declaration
myInt := 10;

# variable declaration with explicit types
bool myBool = true;

# updating a variable
myBool = !myBool;

# variable reassignment is not allowed:
# myInt := 42; -> this will error

{
    int myInt = 42; # this will work, because the variable is declared in a new scope
    myInt = 128; # this will update the variable in the current scope
}

check := myInt == 128; # false

# you can also declare functions

def myFunction() {
    myBool = !myBool;
}

# functions can take arguments
def sum(int a, int b) {
    myInt = a + b;
}

# functions also can return values
def product(int a, int b) int {
    return a * b;
}

def pow(int a, int n) int {
    result := 1;

    # there are also loops

    for i := 0; i < n; i++ {
        result *= a;
    }
    return result;
}

# functions can also call themselves
def fib(int n) int {
    if n < 2 {
        return n;
    }
    return fib(n - 1) + fib(n - 2);
}

# functions currently aren't first class citizens, you can't use them as variables

# while loops

a := 0;
while a < 10 {
    a++;
}

# while loops, just like if, have else blocks, which execute once if the initial condition is false:

b := 0;
while b > 0 {
    b--;
} else {
    b = 42;
}

# nested functions are not allowed

def outerFunction() {
#   def innerFunction() {} -> will raise an error
}

# you can use break and continue keywords in loops
count := 0;
for i := 0; i < 10; i++ {
    if i == 7 {
        break;
    } else {
        if i == 5 {
            continue;
        }
    }
    count++;
}

### Binary operators:

# addition                  +
# subtraction               -
# multiplication            *
# division                  /
# modulo                    %
# or                        ||
# and                       &&
# greater than              >
# less than                 <
# greater than or equal to  >=
# less than or equal to     <=
# strict equality           ==
# strict inequality         !=

### Unary operators

# negation                  !
```
