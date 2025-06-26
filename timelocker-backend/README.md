# TimeLocker Backend

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
- **资产管理**: 基于Covalent API的多链资产余额查询和实时价格更新
- **数据缓存**: Redis缓存提升资产查询性能
- **API文档**: 完整的Swagger文档支持

## 技术栈

- **后端框架**: Gin (Go 1.23.10+)
- **数据库**: PostgreSQL + GORM
- **缓存**: Redis
- **区块链**: go-ethereum
- **认证**: JWT (golang-jwt/jwt)
- **配置**: Viper
- **API文档**: Swagger
- **日志**: Zap
- **资产数据**: Covalent API

## 快速开始

### 1. 环境要求

- Go 1.23.10+
- PostgreSQL 14+
- Redis 6+

### 2. 安装依赖

```bash
go mod tidy
```

### 3. 配置数据库

#### 初始化数据库

创建PostgreSQL数据库：

```sql
CREATE USER timelocker WITH ENCRYPTED PASSWORD 'timelocker';
CREATE DATABASE timelocker_db OWNER timelocker;
GRANT ALL PRIVILEGES ON DATABASE timelocker_db TO timelocker;
```

运行数据库初始化脚本：

```bash
# 进入数据库目录
cd pkg/database

# 运行初始化脚本
./init_db.sh
# 或者手动执行SQL文件
psql -h localhost -U timelocker -d timelocker_db -f init.sql
```

数据库表结构:

```sql
-- 用户表
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    wallet_address VARCHAR(42) NOT NULL UNIQUE,
    chain_id INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_login TIMESTAMP WITH TIME ZONE,
    status INTEGER DEFAULT 1
);

-- 支持的区块链表
CREATE TABLE support_chains (
    id BIGSERIAL PRIMARY KEY,
    chain_name VARCHAR(50) NOT NULL UNIQUE,
    display_name VARCHAR(100) NOT NULL,
    chain_id BIGINT NOT NULL,
    native_token VARCHAR(10) NOT NULL,
    logo_url TEXT,
    is_testnet BOOLEAN NOT NULL DEFAULT false,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 用户资产表
CREATE TABLE user_assets (
    id BIGSERIAL PRIMARY KEY,
    wallet_address VARCHAR(42) NOT NULL REFERENCES users(wallet_address) ON DELETE CASCADE,
    chain_name VARCHAR(50) NOT NULL,
    contract_address VARCHAR(42) NOT NULL DEFAULT '',
    token_symbol VARCHAR(20) NOT NULL,
    token_name VARCHAR(100) NOT NULL,
    token_decimals INTEGER NOT NULL DEFAULT 18,
    balance VARCHAR(100) NOT NULL DEFAULT '0',
    balance_wei VARCHAR(100) NOT NULL DEFAULT '0',
    usd_value DECIMAL(20,8) DEFAULT 0,
    token_price DECIMAL(20,8) DEFAULT 0,
    price_change24h DECIMAL(10,4) DEFAULT 0,
    is_native BOOLEAN NOT NULL DEFAULT false,
    token_logo_url TEXT,
    chain_logo_url TEXT,
    last_updated TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(wallet_address, chain_name, contract_address)
);

-- 时间锁合约表
CREATE TABLE timelocks (
    id BIGSERIAL PRIMARY KEY,
    wallet_address VARCHAR(42) NOT NULL REFERENCES users(wallet_address) ON DELETE CASCADE,
    chain_id INTEGER NOT NULL,
    chain_name VARCHAR(50) NOT NULL,
    contract_address VARCHAR(42) NOT NULL,
    standard VARCHAR(20) NOT NULL CHECK (standard IN ('compound', 'openzeppelin')),
    creator_address VARCHAR(42),
    tx_hash VARCHAR(66),
    min_delay BIGINT,
    proposers TEXT,
    executors TEXT,
    admin VARCHAR(42),
    remark VARCHAR(500) DEFAULT '',
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'deleted')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(wallet_address, chain_id, contract_address)
);
```

### 4. 配置应用

创建配置文件 `config.yaml`:

```yaml
server:
  port: "8080"
  mode: "debug"  # debug, release, test

database:
  host: "localhost"
  port: 5432
  user: "timelocker"
  password: "timelocker"
  dbname: "timelocker_db"
  sslmode: "disable"

redis:
  host: "localhost"
  port: 6379
  password: ""
  db: 0

jwt:
  secret: "your-jwt-secret-here"
  access_expiry: "24h"
  refresh_expiry: "168h"

# Covalent API配置
covalent:
  api_key: "your-covalent-api-key"
  base_url: "https://api.covalenthq.com/v1"
  request_timeout: "30s"
  cache_prefix: "asset:"
  cache_expiry: 300
```

### 5. 启动服务

```bash
# 开发模式
go run cmd/server/main.go

# 或者构建后运行
go build -o timelocker-backend cmd/server/main.go
./timelocker-backend
```

服务将在 `http://localhost:8080` 启动。

### 6. 访问API文档

启动服务后，访问Swagger API文档：

```
http://localhost:8080/swagger/index.html
```

## API 文档

### 认证相关接口

#### 1. 钱包连接认证

**POST** `/api/v1/auth/wallet-connect`

通过钱包签名进行用户认证。前端需要先让用户用钱包对特定消息进行签名，然后将签名结果发送到此接口。

**请求体**:

```json
{
  "wallet_address": "0x742d35cc6bf34c7a14b8f6c8a63f8a12345f6789",
  "signature": "0x1234567890abcdef...",
  "message": "TimeLocker Login Nonce: 1234567890",
  "chain_id": 1
}
```

**响应**:
```json
{
  "success": true,
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_at": "2025-01-01T00:00:00Z",
    "user": {
      "id": 1,
      "wallet_address": "0x742d35cc6bf34c7a14b8f6c8a63f8a12345f6789",
      "chain_id": 1,
      "created_at": "2025-01-01T00:00:00Z",
      "last_login": "2025-01-01T00:00:00Z",
      "status": 1
    }
  }
}
```

#### 2. 刷新访问令牌

**POST** `/api/v1/auth/refresh`

使用刷新令牌获取新的访问令牌。

**请求体**:
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

#### 3. 获取用户资料

**GET** `/api/v1/auth/profile`

获取当前认证用户的资料信息。

**请求头**:
```
Authorization: Bearer <access_token>
```

#### 4. 切换链

**POST** `/api/v1/auth/switch-chain`

切换用户的主链。

**请求体**:
```json
{
  "chain_id": 137
}
```

### 时间锁合约接口

#### 1. 检查时间锁状态

**GET** `/api/v1/timelock/status`

检查当前用户是否拥有时间锁合约。

**响应**:
```json
{
  "success": true,
  "data": {
    "has_timelocks": true,
    "timelocks": [
      {
        "id": 1,
        "wallet_address": "0x742d35cc6bf34c7a14b8f6c8a63f8a12345f6789",
        "chain_id": 1,
        "chain_name": "eth-mainnet",
        "contract_address": "0x1234567890abcdef...",
        "standard": "compound",
        "min_delay": 172800,
        "status": "active",
        "created_at": "2025-01-01T00:00:00Z"
      }
    ]
  }
}
```

#### 2. 创建时间锁合约

**POST** `/api/v1/timelock/create`

创建新的时间锁合约记录。

**请求体**:
```json
{
  "chain_id": 1,
  "chain_name": "eth-mainnet",
  "contract_address": "0x1234567890abcdef...",
  "standard": "compound",
  "creator_address": "0x742d35cc6bf34c7a14b8f6c8a63f8a12345f6789",
  "tx_hash": "0xabcdef...",
  "min_delay": 172800,
  "admin": "0x742d35cc6bf34c7a14b8f6c8a63f8a12345f6789",
  "remark": "我的时间锁合约"
}
```

#### 3. 导入时间锁合约

**POST** `/api/v1/timelock/import`

导入已存在的时间锁合约。

**请求体**:
```json
{
  "chain_id": 1,
  "chain_name": "eth-mainnet",
  "contract_address": "0x1234567890abcdef...",
  "standard": "openzeppelin",
  "abi": "[{...}]",
  "remark": "导入的时间锁合约"
}
```

#### 4. 获取时间锁列表

**GET** `/api/v1/timelock/list`

获取用户的时间锁合约列表。

**查询参数**:
- `page`: 页码（默认1）
- `page_size`: 每页大小（默认10）
- `chain_id`: 筛选链ID
- `standard`: 筛选合约标准
- `status`: 筛选状态

#### 5. 获取时间锁详情

**GET** `/api/v1/timelock/:id`

获取指定时间锁合约的详细信息。

#### 6. 更新时间锁

**PUT** `/api/v1/timelock/:id`

更新时间锁合约的备注信息。

#### 7. 删除时间锁

**DELETE** `/api/v1/timelock/:id`

删除时间锁合约记录。

### 资产相关接口

#### 1. 获取用户资产

**GET** `/api/v1/assets`

获取用户在所有支持链上的资产信息。

**响应**:
```json
{
  "success": true,
  "data": {
    "wallet_address": "0x742d35cc6bf34c7a14b8f6c8a63f8a12345f6789",
    "assets": [
      {
        "chain_name": "eth-mainnet",
        "chain_display_name": "Ethereum Mainnet",
        "chain_id": 1,
        "token_symbol": "ETH",
        "token_name": "Ethereum",
        "balance": "1.5",
        "balance_wei": "1500000000000000000",
        "usd_value": 3750.0,
        "token_price": 2500.0,
        "price_change24h": 2.5,
        "is_native": true,
        "last_updated": "2025-01-01T00:00:00Z"
      }
    ],
    "total_usd_value": 3750.0,
    "last_updated": "2025-01-01T00:00:00Z"
  }
}
```

#### 2. 刷新用户资产

**POST** `/api/v1/assets/refresh`

强制刷新用户资产信息。

### 区块链相关接口

#### 1. 获取支持链列表

**GET** `/api/v1/chain/list`

获取所有支持的区块链列表。

**查询参数**:
- `is_testnet`: 筛选测试网/主网
- `is_active`: 筛选激活状态

#### 2. 根据ID获取链信息

**GET** `/api/v1/chain/:id`

根据数据库ID获取链信息。

#### 3. 根据ChainID获取链信息

**GET** `/api/v1/chain/chainid/:chain_id`

根据区块链ChainID获取链信息。

### 系统接口

#### 健康检查

**GET** `/health`

检查服务健康状态。

**响应**:
```json
{
  "status": "ok",
  "service": "timelocker-backend",
  "version": "1.0.0"
}
```

