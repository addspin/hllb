$TTL 3600
@       IN      SOA     ns1.test.ru. admin.test.ru. (
                        2026020901
                        7200
                        3600
                        1209600
                        86400
                        )

@       IN      NS      ns1.test.ru.
@       IN      NS      ns2.test.ru.
@       IN      NS      ns3.test.ru.
@       IN      A       10.13.1.34
www     IN      A       10.13.1.34

admin     IN      A     10.13.1.37                

*.info     IN      A       10.13.1.33

*.msg.admin    IN      A       10.13.1.99

lb     IN      A       10.13.1.36
lb     IN      A       176.125.254.184