# Microservices Architecture

This project demonstrates a simple microservices architecture using Golang and Docker. It includes User, Product, and Order services, with an API Gateway using Kong.

## Prerequisites

- Docker
- Docker Compose

## Setup and Run

1. **Clone the repository:**
   ```sh
   git clone https://github.com/brinton0739/microservicesarchitecture.git
   cd microservicesarchitecture

2. **Build and run the service using Docker Compose**

    ```sh
    docker-compose up --build
    ```
3. **Test Services With Curl**


    **User Services**

    ```sh
    curl -i -X POST http://localhost/user/register -H "Content-Type: application/json" -d '{"username":"user1", "email":"user1@example.com", "password":"password123"}'

    curl -i -X GET http://localhost/user/users

    ```

    **Product Services**

    ```sh
    curl -i -X POST http://localhost/product/product -H "Content-Type: application/json" -d '{"name":"Product1", "description":"is nice", "price":100}'

    curl -i -X GET http://localhost/product/products

    ```
    

    **Order Services**

    ```sh
    curl -i -X POST http://localhost/order/order -H "Content-Type: application/json" -d '{"user_id":1, "product_id":1, "quantity":2, "status":"pending", "total": 22.22}'

    curl -i -X GET http://localhost/order/orders

    ```
    

3. **Testing**

- [User Service](http://localhost/user/register): `http://localhost/user/register`
- [Product Service](http://localhost/product/products): `http://localhost/product/products`
- [Order Service](http://localhost/order/orders): `http://localhost/order/orders`


4. **License**

This project is licensed under the MIT License.



  


