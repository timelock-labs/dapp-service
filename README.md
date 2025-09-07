# TimeLocker Backend V1

The **TimeLocker Backend Service** provides API support for a decentralized timelock management platform.  

## 🚀 Features

- **Wallet Authentication**  
  Ethereum wallet signature-based authentication, eliminating the need for traditional usernames and passwords.  

- **JWT Tokens**  
  Secure access with access and refresh token mechanisms.  

- **User Management**  
  User profiles are automatically created and managed based on wallet addresses.  

- **Timelock Management**  
  - Supports both **Compound** and **OpenZeppelin** timelock standards  
  - Create, import, query, update, and delete timelock contracts  
  - Manage contract states and view detailed information  

- **Multi-Chain Support**  
  Compatible with major networks including Ethereum, Arbitrum, BSC, Polygon, Optimism, and Base.  

- **Data Caching**  
  Redis caching improves asset query performance.  

- **API Documentation**  
  Comprehensive Swagger documentation.  

---

## 🛠 Tech Stack

- **Backend Framework**: Gin (Go 1.23.10+)  
- **Database**: PostgreSQL (14.18+) + GORM  
- **Cache**: Redis (8.0.2+)  
- **Blockchain**: go-ethereum  
- **Authentication**: JWT (golang-jwt/jwt)  
- **Configuration**: Viper  
- **API Documentation**: Swagger  
- **Logging**: Zap  
