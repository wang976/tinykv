package engine_util

/*
An engine is a low-level system for storing key/value pairs locally (without distribution or any transaction support,
etc.). This package contains code for interacting with such engines.

engine 是一个用于在本地存储键值对的底层系统（不涉及分布式或任何事务支持等）。本包包含与这类 engine 交互的代码。

CF means 'column family'. A good description of column families is given in https://github.com/facebook/rocksdb/wiki/Column-Families
(specifically for RocksDB, but the general concepts are universal). In short, a column family is a key namespace.
Multiple column families are usually implemented as almost separate databases. Importantly each column family can be
configured separately. Writes can be made atomic across column families, which cannot be done for separate databases.

CF 表示 'column family'（列族）。关于列族的详细描述见 https://github.com/facebook/rocksdb/wiki/Column-Families
（该文档专门针对 RocksDB，但其中的一般概念是通用的）。简而言之，列族就是一个键的命名空间。
多个列族通常被实现为几乎相互独立的数据库。重要的是，每个列族可以单独配置。
跨列族的写入可以是原子的，而相互独立的数据库之间则无法做到这一点。

engine_util includes the following packages:

engine_util 包含以下包：

* engines: a data structure for keeping engines required by unistore.
* engines：用于保存 unistore 所需 engine 的数据结构。
*
* write_batch: code to batch writes into a single, atomic 'transaction'.
* write_batch：将多个写入批量合并为一个原子 '事务' 的代码。
*
* cf_iterator: code to iterate over a whole column family in badger.
* cf_iterator：在 badger 中遍历整个列族的代码。
*/
