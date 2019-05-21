# weassembly

wegate的各种子插件

为了避免wegate有过多的tcp连接，所以把一些轻量的，简单的插件都集中在本项目中，一同连接

weassembly由众多子模块注册到一个gate模块中，再由gate模块与wegate连接，如下图所示

``` asciiflow
wegate
  ^
  |
  +
 gate             +--+
  ^                  |
  |                  +-->weassembly
  +----+moduleA      |
  |                  |
  +----+moduleB      |
  |                  |
  +----+moduleC   +--+
```
