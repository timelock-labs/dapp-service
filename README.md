# TimeLocker Backend V1

TimeLocker 后端服务，提供去中心化时间锁管理平台的API服务。

## 功能特性

- **钱包认证**: 支持以太坊钱包签名认证，无需传统用户名密码
- **JWT令牌**: 访问令牌和刷新令牌机制，保障API安全访问
- **用户管理**: 基于钱包地址的用户管理，自动创建和资料维护
- **时间锁管理**: 
  - 支持Compound和OpenZeppelin两种时间锁标准
  - 时间锁合约的创建、导入、查询、更新和删除
  - 合约状态管理和详情查询
- **多链支持**: 支持以太坊、Arbitrum、BSC、Polygon、Optimism、Base等主流网络
- **数据缓存**: Redis缓存提升资产查询性能
- **API文档**: 完整的Swagger文档支持

## 技术栈

- **后端框架**: Gin (Go 1.23.10+)
- **数据库**: PostgreSQL(14.18+) + GORM
- **缓存**: Redis(8.0.2+)
- **区块链**: go-ethereum
- **认证**: JWT (golang-jwt/jwt)
- **配置**: Viper
- **API文档**: Swagger
- **日志**: Zap

