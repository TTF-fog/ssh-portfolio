I guess it's beginner-friendly?
====
## Why i enjoy it
- There's a package for everything, which are *usually* installable with minimum fuss 
- works great for ML, with none of the headache of having to compile OpenCV and other such libraries from scratch 
- it's also pretty decent for making scripts and browser automation, which i do occasionally

There's just one problem - it's too ambiguous (and no, i'm not an elitist as my friend has said). When coding, i prefer static types and type-checking, which also gives powerful LSP
support as a byproduct. And yet, despite me having the most experience with Python, i'm often left fumbling in the darkness—Especially when using external libraries.
There's no guarantee that what you're using will work, and it oft dosen't even have the decency to panic properly.
> before i get lynched by the Python squad, i know that this _simplicity_? is meant to make it work for beginners (and it did for me)
> however now i just can't stand it after tasting statically typed goodness

Caveman Summary - Python simple; beginner :) | Python too simple, not fun for Grog to write, leave Grog confused
so in other words, we've gone full circle.
Also, one other thing.
```python
def hi (test: int):
    print(test*3)
hi(3) #works
hi("hello") #also works???
 ```
Yep, your seeing that right. Python **ignores** even explicit type declarations that are mindbogglingly stupid and obtuse.
Now you can fix these issues, but it requires a bunch of extra tools...