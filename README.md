Introduction
============

Keepdrawing is a terminal program to draw dungeons for tabletop roleplaying campaigns, where spatial awareness matter more than details.

Installation
============

Keepdrawing is written in Go, so you must have Go installed.

Keepdrawing relies on goncurses, which relies on ncurses.  So first make sure you have ncurses' development files installed (see your distribution for details) and pkg-config as well.

Then install goncurses:

    $ go get github.com/rthornton128/goncurses

You can also see http://github.com/rthornton128/goncurses for more details on how to install it.

Running
=======

Then simply run the `main.go` file.

However, if you wish to have mouse support for drawing, set the `$TERM` environment variable to `xterm-1002`.  This will make your terminal provide data of updating the mouse movement.

So:

    $ TERM=xterm-1002 go run src/main.go

And you should be able to draw.

Planned features
================

* Actually working
* Ability to save and load maps
* Draw rooms and delete areas
* Add doors, stairs
* Add points of interest, that is a tile that has a marker on it, upon approach will contain a longer description
* Allow several levels
