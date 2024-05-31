# Microservices Architecture

This project demonstrates a simple microservices architecture using Golang and Docker. It includes User, Product, and Order services, with an API Gateway using Kong.

## Prerequisites

- Docker
- Docker Compose

## Setup and Run

1. **Clone the repository:**
   ```sh
   git clone https://github.com/your-username/microservices-architecture.git
   cd microservices-architecture

2. **Build and run the service using Docker Compose**

    ```sh
    docker-compose up
    ```
3. **Register the services with Kong**


    **User Services**

    ```sh
    curl -i -X POST http://localhost:8081/register -H "Content-Type: application/json" -d '{"username":"user1", "email":"user1@example.com", "password":"password123"}'

    curl -i -X GET http://localhost:8081/users

    ```

    **Product Services**

    ```sh
    curl -i -X POST http://localhost:8082/product -H "Content-Type: application/json" -d '{"name":"Product1", "description":"is nice", "price":100}'

    curl -i -X GET http://localhost:8082/products

    ```
    

    **Order Services**

    ```sh
    curl -i -X POST http://localhost:8083/order -H "Content-Type: application/json" -d '{"user_id":1, "product_id":1, "quantity":2, "status":"pending", "total": 22.22}'

    curl -i -X GET http://localhost:8083/orders

    ```
    

3. **Testing**

- [User Service](http://localhost:8081/user/register): `http://localhost:8081/user/register`
- [Product Service](http://localhost:8082/product/products): `http://localhost:8082/product/products`
- [Order Service](http://localhost:8083/order/orders): `http://localhost:8083/order/orders`


4. **License**

This project is licensed under the MIT License.



  


