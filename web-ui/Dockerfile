FROM python:3.11-slim

WORKDIR /app

COPY requirements.txt requirements.txt
RUN pip install --no-cache-dir -r requirements.txt

COPY . .

# grpcurl install
RUN apt-get update && apt-get install -y curl unzip \
 && curl -sL https://github.com/fullstorydev/grpcurl/releases/download/v1.8.9/grpcurl_1.8.9_linux_x86_64.tar.gz \
 | tar -xz -C /usr/local/bin

EXPOSE 8080
CMD ["python", "app.py"]
