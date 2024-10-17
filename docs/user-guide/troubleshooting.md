# Troubleshooting

## How to access the virtual machine of GitHub runner?

To access the virtual machine, the ID of the virtual machine must be known. The ID can be found in GitHub Actions logs. Once the ID is known, the virtual machine can be accessed by finding the network namespace:

```bash
ip netns ls | grep e520c86a3ce850b055ed8f8e
fireactions-bnk-c8-m16-e520c86a3ce850b055ed8f8e (id: 3)
```

Then, find the IP address of the virtual machine:

```bash
ip netns exec fireactions-bnk-c8-m16-e520c86a3ce850b055ed8f8e ip a
1: lo: <LOOPBACK> mtu 65536 qdisc noop state DOWN group default qlen 1000
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
2: eth0@if20202: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue state UP group default qlen 1000
    link/ether ea:32:bb:c6:49:0a brd ff:ff:ff:ff:ff:ff link-netnsid 0
    inet 192.168.129.59/23 brd 192.168.129.255 scope global eth0
       valid_lft forever preferred_lft forever
    inet6 fe80::e832:bbff:fec6:490a/64 scope link
       valid_lft forever preferred_lft forever
3: tap0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc fq_codel state UP group default qlen 1000
    link/ether b6:cd:19:ae:29:b7 brd ff:ff:ff:ff:ff:ff
    inet6 fe80::b4cd:19ff:feae:29b7/64 scope link
       valid_lft forever preferred_lft forever
```

In this case, the assigned IP address is `192.168.129.59`. The virtual machine can be accessed using SSH:

```bash
ssh -l root 192.168.129.59
```

The default password is `fireactions`. This can be changed by creating a custom image. Once logged in, the virtual machine can be managed as a regular Linux machine. Inside you will find `fireactions` service logs:

```bash
root@localhost:~# systemctl status fireactions
● fireactions.service - Fireactions
     Loaded: loaded (/etc/systemd/system/fireactions.service; enabled; vendor preset: enabled)
     Active: active (running) since Fri 2024-09-27 14:54:48 UTC; 2 weeks 5 days ago
<...>
```
