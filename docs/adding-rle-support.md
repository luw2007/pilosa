# 为 Pilosa 添加行程编码支持

- 原文:[https://www.pilosa.com/blog/adding-行程编码-support/](https://www.pilosa.com/blog/adding-行程编码-support/)

- 翻译：[luw2007](https://github.com/luw2007)

Pilosa 基于64位[Roaring Bitmaps](http://roaringbitmap.org)实现，通常被认为是压缩存储+计算任意位集的最佳方法。直到最近，我们的Roaring软件包缺少一个重要的功能：[行程编码（Run-length_encoding）](https://en.wikipedia.org/wiki/Run-length_encoding)！在即将发布的版本中添加了完整的行程编码支持，我们希望分享实现的一些细节。

## Roaring 相关基础
Roaring Bitmaps 是 Daniel Lemire 等人发明的压缩位图索引技术。他们指出，对位图数据使用三种不同的表示可以获得一般数据的出色性能（存储大小和计算速度）。这三个“容器类型”是整数数组，未压缩的位集和行程编码。


以下是使用这组整数的这些容器类型的具体示例：

```
{0,1,2,3,6,7,9,10,14}
```
![行程编码容器示例](https://www.pilosa.com/img/blog/adding-行程编码-support/行程编码-container-example.png)
行程编码容器类型

数组和运行值存储为16位整数，而此小示例的整个位集仅为16位。显然使用`bitset`空间更小，但每个容器类型适用于数据中的不同模式。例如，如果我们想要存储一组32个连续的整数，则 行程编码 表示将是最小的。因为每个容器都有一个关联主键，它存储容器中所有元素共有的高位。

当我们决定构建一个独立的位图索引时，Roaring 主键成为最好的选择。用 Go 实现它，我们使用64位主键来支持更高的基数，而不是参考实现中的32位主键。在我们开始实施之后，行程编码容器被添加到Roaring作为扩展，但作为一个非关键功能，它在我们的路线图上停留了一段时间。除内存存储外，Roaring 还包含[文件存储的完整规范](https://github.com/RoaringBitmap/RoaringFormatSpec)。除了一些微小（二进制不兼容）的差异，我们密切关注官方动态。

## 添加行程编码
如果您熟悉行程编码，这对于博客文章来说可能看起来很简单。行程编码是最简单的压缩技术之一; 用于编码和解码的函数可以用您喜欢的语言的几行编写。Roaring 速度的关键在于两个容器上的任何操作的计算都是在原始容器上完成的，而不是转换任何一个。让我们考虑`AND`当只实现前两种容器类型时（交叉）操作如何工作。对于`A AND B`，有三种情况：`A`和`B`是数组（array-array），`A`和`B`是位集（bitset-bitset），`A`是数组和`B`是 bitset，反之亦然（array-bitset）。其中每个都必须单独实现，因此您开始了解 Roaring 操作如何比简单的概念`AND`操作更复杂。

添加新的行程编码容器类型后，我们需要三个新函数：RLE-RLE，array-RLE，bitset-RLE。这只是为了`AND`操作; 我们还需要三个新功能`OR`。我们还支持非交换差分操作`ANDNOT`，以前需要四个函数（除上述三个之外的 bitset-array ），现在需要9个（array-RLE, RLE-array, bitset-RLE, RLE-bitset, RLE-RLE）。我们`XOR`在并行分支中添加了操作，因此我们将新的行程编码`XOR`函数包含在另外六个中。这17个新功能只是为了支持这四个操作的行程编码，其中许多都是非常重要的。所有这些操作功能总结在下表中。

![行程编码操作功能](https://www.pilosa.com/img/blog/adding-rle-support/rle-function-tables.png)
行程编码操作函数 “x”表示所需的函数; 绿色的是新增函数

在一个行程编码容器上运行的函数往往更复杂，并且在两个行程编码容器上运行的函数更是如此。例如，`intersectRunRun`用于计算`AND`两个行程编码容器的函数同时迭代每个容器中的运行。对于遇到的每对运行，存在六种不同的情况，一种用于两个间隔可以彼此重叠的每种方式。`differenceRunRun`可能是所有操作中最棘手的。同样，必须考虑几种不同的重叠情况，但与交叉算法不同，这些情况是交错的。

而且这还不是全部，除了二进制操作之外，Roaring 还需要做很多其他事情。所有这些操作都需要在行程编码容器上或与行程编码容器一起支持：

* 设置和清除位。
* 写入和读取文件以进行持久存储。
* 转换容器类型以获得最佳存储大小，并决定何时执行此操作。
* 计算容器类型的汇总值：count，runCount，max。其中一些也是非常重要的，[Roaring paper](https://arxiv.org/pdf/1603.06549.pdf)中描述了非常聪明的解决方案。
* 迭代一个可以包含所有三种容器类型的混合位图。
* 内部`intersectionCount`函数可加速某些查询。

当然还有单元测试。Roaring 是 Pilosa 的核心，所以我们尽可能彻底地测试它。行程编码工作包括1500多个新的特征代码行，以及2500多个新的单元测试行。虽然我们的 Roaring 包功能齐全，但我们仍然在 todo 列表上有一些任务：

* 对大型实际数据进行彻底的基准测试和测试。
* 扩展模糊测试。
* 检查反向存储，查找“稀疏零”数据。

如果您想查看详细信息，可以在此[pull request](https://github.com/pilosa/pilosa/pull/661/files)中看到大部分工作，或者仅在当前的[Roaring package](https://github.com/pilosa/pilosa/tree/master/roaring)中看到。我们也很乐意帮助您分享并尝试一下 Pilosa ！

## 偏离 Roaring 规范
为了后期方便：

* 操作支持：在早期，我们只需要一个二进制操作的子集，特别是缺少`XOR`和`NOT`。
* 不兼容的文件规格差异：
	* 我们的“cookie”总是0-3字节; 我们的容器计数总是4-7字节，从不使用2-3字节。这只是简化了编写和读取文件的逻辑。我们的魔法数字与Roaring魔法数字相匹配。
	* 我们的 cookie 包括文件格式版本，以2-3字节为单位（此版本为零）。
	* 我们的偏移标题部分始终包含在内。在规范中，它被省略为小位图，这解释了为什么它没有滚动到偏移/键部分。这是我们不必要的解析复杂性。
	* 行程编码运行被序列化为[start，last]，而不是[start，length]。
* 容器存储部分是一个未指定长度的操作日志。这会保留位图更新的记录，在读取文件时会对其进行处理。有关这方面的更多详情，请继续关注 Pilosa 即将发布的帖子。

我们的文件格式在[文档](https://www.pilosa.com/docs/latest/architecture/#roaring-bitmap-storage-format)中有详细描述。

