
TARGET=xlsoa.example.greeter

$(TARGET): greeter_server.go
	go build -o $@ $^

clean:
	rm -f $(TARGET)
