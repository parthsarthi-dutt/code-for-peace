FROM eclipse-temurin:17-jdk-jammy

RUN useradd -m -u 1001 sandbox

WORKDIR /sandbox

# No network, dropped capabilities are handled at runtime via docker flags
