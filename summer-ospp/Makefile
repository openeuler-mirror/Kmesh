# Makefile

# 定义变量
TARGET = ebpf-based-cni
SOURCES = main.go

# cni agant
AGENT_TARGET = agant
AGENT_SOURCES = ./httpserver/server.go

# 定义编译器和编译选项
GO = go
GOFLAGS =

all: $(TARGET)

# 编译目标
$(TARGET): $(SOURCES)
	$(GO) build $(GOFLAGS) -o /opt/cni/bin/$(TARGET) $(SOURCES)
	$(GO) build $(GOFLAGS) -o /root/summer-ospp/bin/$(AGENT_TARGET) $(AGENT_SOURCES)
	# scp /root/summer-ospp/bin/$(AGENT_TARGET) root@192.168.43.158:/root/summer-ospp/bin/$(AGENT_TARGET)
	# scp /opt/cni/bin/$(TARGET) root@192.168.43.158:/opt/cni/bin/$(TARGET)

# 清理编译产物
clean:
	rm -f /opt/cni/bin/$(TARGET)
	rm -f /root/summer-ospp/bin/$(AGENT_TARGET)
