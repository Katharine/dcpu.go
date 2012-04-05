This is a direct copy of the [comment from Reddit][reddit] that it was created for.

-------

As an addendum to the previous comment, I went and [hacked an emulator together for you](https://github.com/Katharine/dcpu.go). It's written in Go and is all of 340 lines in total, including tests.

Some notes:

* It's a rush job. Not a great coding masterpiece.
* The only test is that the sample program given has the expected value in register X when it ends (it does).
* Since programs cannot exit or receive input or output, it exposes an interface for stepping through and direct manipulation of memory and registers.
* Since most instructions take a value and then store the result in the same place it uses lots of pointers.

Perhaps it will be of some interest to you, if you are actually interested in how one might go about making such things.

The general principle is that, having loaded in a memory image (using `LoadFile` or `LoadStream`, you call `Run` (or `ExecuteCycle` in a loop, which is the same).

`ExecuteCycle` reads one word out of memory and increments the program counter. It decodes the word as specified in [the spec](http://0x10c.com/doc/dcpu-16.txt), uses `resolve` to look up the actual values being passed along with the opcode, then looks up the opcode in the `basicOpcodes` array or `extendedOpcodes` map.

The only other interesting point is that some operations require skipping the next one. These operations set `skipping` to `true`, in which case no values are looked up and no operations performed on the next round, which sets it to `false` again.

That's about it. The documentation, such as I actually included any, can be [viewed here](http://gopkgdoc.appspot.com/pkg/github.com/Katharine/dcpu.go).

[reddit]: http://www.reddit.com/r/learnprogramming/comments/rtpkq/how_can_i_write_a_dcpu_assembler_and_emulator_in/c48n3l6
