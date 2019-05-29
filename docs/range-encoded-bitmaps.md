# 使用位图执行范围查询

- 原文:[https://www.pilosa.com/blog/range-encoded-bitmaps/](https://www.pilosa.com/blog/range-encoded-bitmaps/)

- 翻译: luw2007

Pilosa 在 Base-2，Range-Encoded，Bit-slices 索引中存储整数值。这篇文章解释了怎么做到的。

## Introduction 介绍 

Pilosa 的理念是：将所有内容表示为位图。例如，对象之间的关系被视为布尔值（关系存在或不存在）并且这些布尔值的集合被存储为一系列1和0。这些布尔集是位图，通过在各种位运算中使用位图，我们可以非常快速地执行复杂查询。此外，位图压缩技术使我们能够以非常紧凑的格式表示大量数据，从而减少了存储数据和执行查询所需的资源量。

使用位图和 roaring 位图压缩[roaring bitmap compression](http://roaringbitmap.org/)，Pilosa 非常擅长对数十亿个对象执行分段查询。但是我们经常会看到使用整数值有用的用例。例如，我们可能希望查询`foo`大于 1,000 的所有记录。这篇文章逐步解释了我们如何将范围编码的位图添加到 Pilosa，使我们能够在查询操作中支持整数值。

本文中描述的所有概念都是基于过去几十年来一些非常聪明的人所做的研究。这是我尝试在更高层次上描述事物，但我鼓励您阅读有关[比特切片索引](https://cs.brown.edu/courses/cs227/archives/2008/Papers/Indexing/buchmann98.pdf)和[范围编码](https://link.springer.com/chapter/10.1007/3-540-45675-9_8)的更多信息。

## 位图编码(Bitmap Encoding)

首先，让我们假设我们想要以这样一种方式对[动物王国](https://en.wikipedia.org/wiki/Animal)的每个成员进行编目，以便我们可以根据他们的特征轻松有效地探索各种物种。因为我们讨论的是位图，所以示例数据集可能如下所示：

![示例数据集](https://www.pilosa.com/img/blog/range-encoded-bitmaps/example-dataset.png) 示例数据集

### 等式编码的位图(Equality-encoded Bitmaps)

上面的示例显示了一组等式编码的位图，其中每一行(每个特征)是一个位图，指示哪些动物具有该特征。虽然这是一个相当简单的概念，但是相等编码(equality-encoding)可能非常强大。因为它允许我们将所有东西都表示为布尔关系（即海牛 manatee 有翅膀：是/否），我们可以对数据执行各种按位操作。

下图显示了我们如何通过对位图 `Invertebrate`（无脊椎动物）和位图 `Breathes Air`（呼吸空气）执行逻辑`AND`来找到所有呼吸空气的无脊椎动物。根据我们的样本数据，我们可以看到香蕉蛞蝓 Banana Slug，小灰蜗牛 Garden Snail和轮背猎蝽 Wheel Bug都具有这两个特征。

![按位交点示例](https://www.pilosa.com/img/blog/range-encoded-bitmaps/bitwise-intersection.png) 两个特征的逐位交集

使用等式编码的位图可以获得相当好的结果，但是布尔值不能最好地表示原始数据的情况呢？如果我们想要添加一个特征`Captivity`，该特征标示当前在`Captivity`中持有的样本总数，我们想要执行由这些值过滤的查询？（正如您可能怀疑的那样，我们`Captivity`在下面的示例中使用的值是完全组成的，但它们有助于演示这些概念。）

鉴于我们对平等编码位图的了解，我们可以采用一些（公认的常规）方法。一种方法是为每个可能的值创建`Captivity`一个特征（位图），如下所示：

![Captivity计为单个位图](https://www.pilosa.com/img/blog/range-encoded-bitmaps/captive-rows.png) Captivity Counts表示为单个位图

这种方法没问题，但它有一些限制。首先，它效率不高。根据基数，您可能需要创建大量位图来表示所有可能的值。其次，如果要按一系列`Captivity`值过滤查询，则必须对范围内的每个可能值执行`OR`操作。为了知道哪些动物的`Captivity`数量少于100个，您的查询需要执行类似的操作（Captivity=99 OR Captivity=98 OR Captivity=97 OR ...）。明白了吗。

另一种方法是创建`Captivity`范围桶，而不是将每个可能的值表示为唯一的位图。在这种情况下，您可能会有这样的事情：

![`Captivity`计数为桶](https://www.pilosa.com/img/blog/range-encoded-bitmaps/captive-buckets.png) `Captivity`计数表示为桶

这种方法的一个好处是它的效率更高一些。它也更容易查询，因为您不必构造多个位图的联合以表示一系列值。缺点是它不是那么精细; 通过转换`47`到0-99桶，您正在丢失信息。

这些方法中的任何一种都是对某些问题的完全有效的解决方案，但是对于基数极高且丢失信息是不可接受的情况，我们需要另一种方法来表示非布尔值。我们需要这样做，以便我们可以对值范围执行查询，而无需编写非常大且繁琐的`OR`操作。为此，我们来谈谈范围编码的位图以及它们如何避免我们在之前的方法中遇到的一些问题。

### 范围编码位图(Range-Encoded Bitmaps)
首先，让我们看一下上面的例子，看看我们使用范围编码位图时的样子。

![`Captivity`计数为范围编码的位图](https://www.pilosa.com/img/blog/range-encoded-bitmaps/captive-range-encoded-rows.png)`Captivity`计数为范围编码的位图

使用范围编码位图表示值与使用相等编码执行的操作类似，但不是仅设置与特定值对应的位，而是为每个大于实际值的值设置一个位。例如，因为有14个考拉 Koala Bears的灵敏度，我们在位图14中设置位以及位图15,16,17等。而不是代表具有特定`Captivity`计数的所有动物的位图，现在位图代表所有被`Captivity`的动物数量达到并包括该数量。

这种编码方法允许我们执行之前执行的那些范围查询，但不是`OR`在许多不同的位图上执行操作，我们可以从一个或两个位图获得我们想要的内容。例如，如果我们想要知道哪些动物的`Captivity`数量少于15个，我们只需拉出14位图即可完成。如果我们想知道哪些动物的`Captivity`超过15个，那就更复杂了，但并不多。为此，我们拉出表示最大计数的位图（在我们的例子中是956位图），然后减去15位图。

这些操作比我们以前的方法简单得多，而且效率更高。`OR`为了找到我们的范围，我们已经解决了让我们将数十个位图组合在一起的问题，并且我们没有像在分组方法中那样丢失任何信息。但是我们仍然有一些问题使得这种方法不够理想。首先，我们仍然需要保留一个代表每个特定`Captivity`计数的位图。最重要的是，我们增加了复杂性和开销，不仅要为我们感兴趣的值设置一点，而且还要为每个大于该值的值设置一点。这很可能会在大量使用的情况下引入性能问题。

理想情况下，我们想要的是具有范围编码位图的功能和相等编码的效率。接下来，我们将讨论位切片索引，看看它如何帮助我们实现我们想要的目标。

### 位切片索引(Bit-sliced Indexes)
如果我们想使用范围编码位图表示0到956之间的每个可能值，我们必须使用957位图。虽然这有效，但它不是最有效的方法，当可能值的基数变得非常高时，我们需要维护的位图数量会变得过高。位切片索引让我们以更有效的方式表示这些相同的值。

让我们看一下我们的示例数据，并讨论如何使用位切片索引来表示它。

![`Captivity` 计数为Base-10的位切片索引](https://www.pilosa.com/img/blog/range-encoded-bitmaps/captive-bsi-base10.png) Base-10的位切片索引

请注意，我们已将值分解为三个Base-10的组。第一列位表示值003，即`Captivity`中的海牛数。组0 `003`是`3`，所以我们在组0，行3中设置了一个位。组1和组2 `003`都是0，所以我们在组1，第0行和组2，第0行中设置位。我们的Base-10的索引中的每个组需要10个位图来表示所有可能的值，因此在我们需要表示0到956范围内的值的`Captivity`示例中，我们只需要（3 x 10）= 30位图（而不是我们使用时需要的957位图）每个不同值的位图）。

这很好，但我们基本上只是找到了一种通过我们的平等编码策略提高效率的方法。让我们看看当我们将位片索引与范围编码相结合时的样子。

## 范围编码的位片索引(Range-Encoded Bit-Slice Indexes)
![`Captivity` 计数为范围编码，Base-10的位切片索引](https://www.pilosa.com/img/blog/range-encoded-bitmaps/captive-bsi-range-encoded-base10.png) 范围编码，Base-10，位切片索引

请注意，每个组中最重要的值（Base-10的情况下为9）始终为1。因此，我们不需要存储最高价值。因此，对于Base-10，范围编码的位切片索引，我们只需要9个位图来表示组。除此之外，我们还需要存储一个名为“Not Null”的位图，它指示是否为该列设置了值。下图显示了生成的位图。

[`Captivity` 计数为范围编码，Base-10的位切片索引，具有非空](https://www.pilosa.com/img/blog/range-encoded-bitmaps/captive-bsi-range-encoded-base10-not-null.png) 范围编码，Base-10，位切片索引，Not-Null

因此，对于3分量值，我们需要（（3 x 9）+ 1）= 28位图来表示0到999范围内的任何值。现在我们有一种非常有效的方法来存储值，我们得到了范围的好处编码，因此我们可以执行过滤范围的查询。让我们更进一步，尝试编码我们的值范围的Base-2表示。

### Base-2
我们如果`Captivity`不使用Base-2表示我们的值，而是使用Base-2，那么我们最终得到一组范围编码的位切片索引，如下所示：

![容量计数为范围编码的Base-2位切片索引](https://www.pilosa.com/img/blog/range-encoded-bitmaps/captive-bsi-range-encoded-base2.png) 范围编码，Base-2，位切片索引

第一列位代表Base-2值`000000011`，它是`Captivity`中的海牛Manatees数量（3 in Base-10）。由于组0和组1 000000011都是1，我们在组0，行1和组1，行1中设置了一个位。由于其余组`000000011`都是0，所以我们在第0行为组2到9设置了一个位，并且（因为这些是范围编码的）我们在每个大于0的值中设置一个位。在Base-2的情况下，这意味着我们还在第1行中为组2到9设置位。

但是请记住，就像我们之前看到的base-10表示的位图9一样，位图1总是一个，所以我们不需要存储它。这让我们留下了这个：

![`Captivity` 计数为范围编码，Base-2，位切片索引，具有非空](https://www.pilosa.com/img/blog/range-encoded-bitmaps/captive-bsi-range-encoded-base2-not-null.png) 具有非空的范围编码，Base-2，位切片索引

通过这种编码，我们可以只用10位图来表示样本值的范围！另外，请注意Base-2，范围编码的位切片索引是整数值的二进制表示的倒数。这告诉我们的是，我们可以仅使用（n + 1）个位图（其中附加位图是“Not Null”位图）来表示基数为n的任何值范围。这意味着我们可以对大整数值执行范围查询，而无需存储不合理数量的位图。

## Pilosa中的范围编码位图
通过在 Pilosa 中实现范围编码位图，用户现在可以存储与数十亿对象相关的整数值，并且可以非常快速地执行按一系列值过滤的查询。我们还支持聚合查询`Sum()`。被`Captivity`的有翼脊椎动物总数是多少？没问题。

作为最后一个练习，让我们演示如何`Captivity`在 Pilosa 中存储和查询我们的示例数据。

```bash
# Create an index called "animals".
curl -X POST localhost:10101/index/animals

# Create a frame "traits" to hold captivity values.
curl localhost:10101/index/animals/frame/traits \
  -X POST \
  -d '{"options":{"rangeEnabled": true,
                     "fields": [{"name": "captivity",
                                 "type": "int",
                                 "min": 0,
                                 "max": 956}]
                    }
         }'

# Add the captivity values to the field.
curl localhost:10101/index/animals/query \
  -X POST \
  -d 'SetFieldValue(frame=traits, col=1,  captivity=3)
      SetFieldValue(frame=traits, col=2,  captivity=392)
      SetFieldValue(frame=traits, col=3,  captivity=47)
      SetFieldValue(frame=traits, col=4,  captivity=956)
      SetFieldValue(frame=traits, col=5,  captivity=219)
      SetFieldValue(frame=traits, col=6,  captivity=14)
      SetFieldValue(frame=traits, col=7,  captivity=47)
      SetFieldValue(frame=traits, col=8,  captivity=504)
      SetFieldValue(frame=traits, col=9,  captivity=21)
      SetFieldValue(frame=traits, col=10, captivity=0)
      SetFieldValue(frame=traits, col=11, captivity=123)
      SetFieldValue(frame=traits, col=12, captivity=318)
  '

# Query for all animals with more than 100 specimens
# in captivity.
curl localhost:10101/index/animals/query \
  -X POST \
  -d 'Range(frame=traits, captivity > 100)'
 
# Query for the total number of animals in captivity
curl localhost:10101/index/animals/query \
  -X POST \
  -d 'Sum(frame=traits, field=captivity)'
```
##结论
在这篇文章中描述的示例显示了我们如何使用位切片索引来显着减少表示一系列整数值所需的位图数量。通过将范围编码应用于索引，我们可以对数据执行各种范围查询。下图比较了我们讨论的不同方法。

![比较图表](https://www.pilosa.com/img/blog/range-encoded-bitmaps/example-comparison.png) 比较图表

我们在版本0.7.0中添加了Range-Encoding支持。您还应该查看范围编码文档。

尝试一下，让我们知道您的想法。我们一直在寻求改进并感谢您的任何反馈！

Travis 是 Pilosa 的首席架构师。 在 Twitter 上找到他 [@travislturner](https://twitter.com/travislturner?lang=en)。
