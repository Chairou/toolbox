#include <stdio.h>
#include <stdlib.h>
#include "libecc.h"

int main() {
    printf("driver for libecc\n");
    struct ecc_generate_keys_return keys = ecc_generate_keys(256);
    printf("keys.r0: %d\n",keys.r0);

    char* pub = ecc_pub_pem(keys.r0);
    //char n = *(pem);
    //size_t length = sizeof(&(pem[0]))/sizeof(char); // pointer sizeof, not array sizeof
    printf("%s\n", pub);
    char* priv = ecc_priv_pem(keys.r1,"abc");
    printf("%s\n", priv);

    printf("decode PEMs to new ids\n");
    printf("pubID: %d\n", ecc_pub_decode_pem(pub));
    printf("privID: %d\n", ecc_priv_decode_pem(priv, "abc"));
    
    printf("ecc_pub_marshal()\n");
    ecc_bytes* b = ecc_pub_marshal(keys.r0);
    printf("marshal len(%d)\n", b->n);
    for(size_t i = 0; i < b->n; i++)
    {
        printf("%02x", b->data[i]);
    }
    printf("\n");
    
    printf("exit");
    return 0;
}