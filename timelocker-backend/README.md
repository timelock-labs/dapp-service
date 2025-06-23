# TimeLocker Backend

TimeLocker 后端服务，提供去中心化时间锁管理平台的API服务。

## 功能特性

- **钱包认证**: 支持以太坊钱包签名认证
- **JWT令牌**: 访问令牌和刷新令牌机制
- **用户管理**: 自动用户创建和资料管理
- **多链支持**: 支持以太坊、Arbitrum、BSC、Polygon、Optimism、Base等网络
- **资产管理**: 多链资产余额查询和监控
- **价格服务**: 实时代币价格更新和缓存
- **智能合约交互**: 通过RPC客户端与区块链交互
- **数据缓存**: Redis缓存提升性能

## 技术栈

- **后端框架**: Gin (Go)
- **数据库**: PostgreSQL + GORM
- **缓存**: Redis
- **区块链**: go-ethereum
- **认证**: JWT
- **配置**: Viper
- **API文档**: Swagger
- **日志**: Zap
- **RPC**: Alchemy , Infura
- **价格数据**: CoinGecko API



## 快速开始

### 1. 环境要求

- Go 1.23.10+
- PostgreSQL 14.18+
- Redis 8.0.2+

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

# 运行初始化脚本（包含表创建和初始数据）
./init_db.sh
# 或者手动执行SQL文件
psql -h localhost -U timelocker -d timelocker_db -f init.sql
```

数据库表结构:

```sql
-- 用户表
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    wallet_address VARCHAR(42) NOT NULL,
    chain_id INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    last_login TIMESTAMP,
    preferences JSONB DEFAULT '{}',
    status INTEGER DEFAULT 1,
    UNIQUE(wallet_address, chain_id)
);

-- 支持的区块链表
CREATE TABLE support_chains (
    id BIGSERIAL PRIMARY KEY,
    chain_id BIGINT NOT NULL UNIQUE,
    name VARCHAR(50) NOT NULL,
    symbol VARCHAR(10) NOT NULL,
    rpc_provider VARCHAR(20) NOT NULL DEFAULT 'alchemy',
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 支持的代币表
CREATE TABLE support_tokens (
    id BIGSERIAL PRIMARY KEY,
    symbol VARCHAR(10) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    coingecko_id VARCHAR(50) NOT NULL UNIQUE,
    decimals INTEGER NOT NULL DEFAULT 18,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 链代币关联表
CREATE TABLE chain_tokens (
    id BIGSERIAL PRIMARY KEY,
    chain_id BIGINT NOT NULL REFERENCES support_chains(id),
    token_id BIGINT NOT NULL REFERENCES support_tokens(id),
    contract_address VARCHAR(42) DEFAULT '',
    is_native BOOLEAN NOT NULL DEFAULT false,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(chain_id, token_id)
);

-- 用户资产表
CREATE TABLE user_assets (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    wallet_address VARCHAR(42) NOT NULL,
    chain_id BIGINT NOT NULL,
    token_id BIGINT NOT NULL REFERENCES support_tokens(id),
    balance VARCHAR(100) NOT NULL DEFAULT '0',
    balance_wei VARCHAR(100) NOT NULL DEFAULT '0',
    usd_value DECIMAL(20,8) DEFAULT 0,
    last_updated TIMESTAMP DEFAULT NOW(),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(user_id, chain_id, token_id)
);
```

#### 添加新的区块链支持

1. 在数据库中添加链配置:

```sql
INSERT INTO support_chains (chain_id, name, symbol, rpc_provider, is_active) 
VALUES (your_chain_id, 'Chain Name', 'SYMBOL', 'alchemy', true);
```

2. 在配置文件中添加RPC配置
3. 在 `blockchain/rpc_client.go` 中添加连接逻辑
4. 添加对应的代币配置

#### 添加新的代币支持

1. 在数据库中添加代币:

```sql
INSERT INTO support_tokens (symbol, name, coingecko_id, decimals, is_active) 
VALUES ('TOKEN', 'Token Name', 'coingecko-id', 18, true);
```

2. 添加链代币关联:

```sql
INSERT INTO chain_tokens (chain_id, token_id, contract_address, is_native, is_active) 
VALUES (chain_id, token_id, 'contract_address', false, true);
```

### 4. 配置应用

配置文件 `config.yaml`

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

通过钱包签名进行用户认证。前端需要先让用户签名一个消息，然后将签名结果发送到此接口。

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
      "preferences": {},
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

**响应**:
```json
{
  "success": true,
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_at": "2025-01-01T00:00:00Z"
  }
}
```

#### 3. 获取用户资料

**GET** `/api/v1/auth/profile`

获取当前认证用户的资料信息。

**请求头**:
```
Authorization: Bearer <access_token>
```

**响应**:
```json
{
  "success": true,
  "data": {
    "id": 1,
    "wallet_address": "0x742d35cc6bf34c7a14b8f6c8a63f8a12345f6789",
    "chain_id": 1,
    "created_at": "2025-01-01T00:00:00Z",
    "last_login": "2025-01-01T00:00:00Z",
    "preferences": {},
    "status": 1
  }
}
```

#### 4. 切换链

**POST** `/api/v1/auth/switch-chain`

切换用户的主链。

**请求头**:
```
Authorization: Bearer <access_token>
```

**请求体**:
```json
{
  "chain_id": 12
}
```

### 资产相关接口

#### 1. 获取用户资产

**GET** `/api/v1/assets`

获取用户在所有链上的资产信息。

**请求头**:
```
Authorization: Bearer <access_token>
```

**查询参数**:
- `chain_id` (可选): 主要显示的链ID
- `force_refresh` (可选): 是否强制刷新，默认false

**响应**:
```json
{
  "success": true,
  "data": {
    "wallet_address": "0x742d35cc6bf34c7a14b8f6c8a63f8a12345f6789",
    "primary_chain_id": 1,
    "primary_chain": {
      "chain_id": 1,
      "chain_name": "Ethereum",
      "chain_symbol": "ETH",
      "assets": [
        {
          "token_symbol": "ETH",
          "token_name": "Ethereum",
          "balance": "1.5",
          "balance_wei": "1500000000000000000",
          "usd_value": 3750.0,
          "token_price": 2500.0,
          "change_24h": 2.5,
          "is_native": true,
          "last_updated": "2025-01-01T00:00:00Z"
        },
        {
          "token_symbol": "USDC",
          "token_name": "USD Coin",
          "contract_address": "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48",
          "balance": "1000.0",
          "balance_wei": "1000000000",
          "usd_value": 1000.0,
          "token_price": 1.0,
          "change_24h": 0.1,
          "is_native": false,
          "last_updated": "2025-01-01T00:00:00Z"
        }
      ],
      "total_usd_value": 4750.0,
      "last_updated": "2025-01-01T00:00:00Z"
    },
    "other_chains": [],
    "total_usd_value": 5950.0,
    "last_updated": "2025-01-01T00:00:00Z"
  }
}
```

#### 2. 刷新用户资产

**POST** `/api/v1/assets/refresh`

强制刷新用户在指定链上的资产信息。

**请求头**:
```
Authorization: Bearer <access_token>
```

**请求体**:
```json
{
  "wallet_address": "0x742d35cc6bf34c7a14b8f6c8a63f8a12345f6789",
  "chain_id": 1,
  "force_refresh": true
}
```

**响应**:
```json
{
  "success": true,
  "data": "Assets refresh initiated successfully"
}
```

### 系统接口

#### 健康检查

**GET** `/health`

检查服务健康状态。

**响应**:
```json
{
  "status": "ok",
  "service": "timelocker-backend"
}
```

