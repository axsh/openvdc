# Editing these

These illustrations were made using [draw.io](http://draw.io/). They can be imported directly and all connections and other metadata should be in place.

### Before you commit

* Export as .svg and set a borderwidth of 10.

* Save the resulting .svg locally and run the following command on it.

```
xmllint --format /path/to/new.svg > /path/to/toyocho-tools/illustrations/new.svg
```

#### But why?

* The border is because the line near the edge of the physical illustration will become very hard to see otherwise.

* The `xmllint` command is used because draw.io saves .svg files' entire XML on one line. This is great for browsers but not great for github. This way git would create a new copy of the entire drawing for every tiny change. Not good for repo size. ;) `xmllint` makes the .svg file multi-line so git only needs to replace the lines that actually changed.
