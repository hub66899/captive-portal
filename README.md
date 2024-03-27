# Captive Portal
為OpenWRT提供網路接入授權

## 在firewall中創建set
類型為mac地址 

默認名字allowed_mac 

timeout為24小時

## 為dns創建port forward規則
規則為源地址mac不在set中的udp53映射到captive portal的dns端口(1088)