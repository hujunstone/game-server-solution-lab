
# DB 
官方推荐是 CockroachDB
pg sql 也可以用 传统关系型数据库，核心功能基于单节点设计，需通过外部工具（如Citus）实现分布式扩展
CockroachDB 可扩展性 支持分布式事务 ，原生支持多节点水平扩展，数据自动分片和复制

1. 如何配置本地环境把这个开源工程run起来
A 用 docker 跑 docker-compose 
B 

2. 参考这个需求，选择一个简单的功能二次开发调用这个工程

配置本地环境把这个开源工程run起来，参考这个需求，选择一个简单的功能二次开发使用这个工程

3. 看下怎么用go和lua分别调用这个库

建一个另外的工程跑这个测试和验证
game-server-solution-lab


cockroachDB不要用8080端口


