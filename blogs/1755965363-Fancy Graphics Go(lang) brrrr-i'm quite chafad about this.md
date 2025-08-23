We have Images Now!*
===== 

#### *T&C May Apply

<mdimg  "images/thumbs.png-80-80" >
i'm sorry i just couldn't resist

there's also categories for blogs now (take a look downstairs)

# Here's the story of how I got images working over SSH
Ever since i first had this idea, there was always one stumbling block - Images. Inside the terminal, rendering images is a messy effort—There's no standardized way, or even one that works everywhere while still being approximate.
It's even more challenging when you get into TUIs. Few protocols mesh well, and even fewer frameworks are supported
> There's only one PR for bubbletea, and it died a year ago

# So, i was left with 4 (equally terrible options) -
- ASCII Blocks
  - These are by far the worst. you don't get color, or even any information. The only benefit is compatibility
- Kitty Graphics Protocol
  - It has the best features, but the worst support, with only kitty having it by default
- Sixel
  - Probably the best, but i couldn't find a good implementation in Go
- Tomfoolery 
  - the less said about this the better

I was initially leaning towards Sixel, but i only found [one](https://github.com/BourgeoisBear/rasterm) library, with no documentation
that's when i found it - 
[Chafa](https://hpjansson.org/chafa/). Not only was there a great library for Go too, it ran nearly everywhere, WITH redundant modes built in (also it looked amazing)

