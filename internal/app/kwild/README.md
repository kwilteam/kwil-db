# kwild
The grpc server for the kwild service.

## Overview
It has the following services:
- [ConfigService](#configservice)
- [TxService](#txservice)
- [PricingService](#pricingservice)
- [AccountService](#accountservice)
- [HealthService](#healthservice)

### ConfigService
SDK/client could use this service to get the configuration of the kwil service.

### TxService
This service handles the core business logic, including:
- Create a database
- Delete a database
- Update a table

### PricingService
This service handles the pricing logic. As blockchain requires computation and storage, the pricing is very important.
The pricing service will calculate the price for each transaction.

### AccountService
This service handles the account logic. It will check the account balance and the account status.

### HealthService
This implements a grpc health check service.