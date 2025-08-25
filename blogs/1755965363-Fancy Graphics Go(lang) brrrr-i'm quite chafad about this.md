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
[Chafa](https://hpjansson.org/chafa/). Not only was there a great library for Go too, it ran nearly everywhere, WITH redundant modes built in (also it fitted amazingly with the look i was going for)

# The Implementation
one good thing about Chafa, is that there are utilities for displaying it directly as a string. This allows me to just stuff it into the application
without too much extra work. I got to work and quickly built a parser for Markdown to insert the images

however, i quickly ran into a problem. Glamour, which i use to render Markdown threw a fit every single time i inserted an image and b
lurted out ASCII control codes like crazy. The workaround was to insert placeholders, and then later render it into text.

```go
	images = append(images, getImageString(parsed[0], int32(width), int32(height), FONT_WIDTH, FONT_HEIGHT))
		newMarkdown += md[:tagstart] + fmt.Sprintf("PLACE%dHOLDER", IND) + md[tagend:] + "\n"

	}
	return newMarkdown, images
```
this allowed me to do the Markdown parsing first, and then the image insertion later
```go
func parseMarkdownAgainForImages(md string, Images []string, imageStyle lipgloss.Style, wrapWidth int) string {
	renderer, _ := glamour.NewTermRenderer(glamour.WithStandardStyle("dark"), glamour.WithWordWrap(wrapWidth))
	str, _ := renderer.Render(md)
	for i, imgData := range Images {
		placeholder := fmt.Sprintf("PLACE%dHOLDER", i)
		str = strings.Replace(str, placeholder, imageStyle.Render(imgData), 1)
	}
	return str
}

```
A couple of bugfixes later, and we have images!!!


<mdimg  "images/img_3.png-120-120" >

(ok the resolution isn't the best but you get my point)