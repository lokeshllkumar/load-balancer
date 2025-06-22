# OpenJDK base image
FROM openjdk:17-jdk-slim

# set the working directory
WORKDIR /app

# copying prebuilt JAR file into the container
COPY target/*.jar app.jar

# exposing ports for access

# port for HTTP/REST
EXPOSE 8081
# port for gRPC
EXPOSE 9091

ENTRYPOINT ["java","-jar","app.jar"]
