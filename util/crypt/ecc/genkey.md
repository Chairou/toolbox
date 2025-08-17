# 生成 secp521r1 私钥
openssl ecparam -genkey -name secp521r1 -out private.pem

# 提取公钥
openssl ec -in private.pem -pubout -out public.pem

# 加密私钥
openssl ec -aes256 -in private.pem -out encrypted_private.pem

# 转换私钥为 PKCS#8 (不可用)
# openssl pkcs8 -topk8 -in private.pem -out secure_private.pem -v2 aes-256-cbc 
