# iptables

In order to route all outgoing traffic to the sidecar, [iptables](https://linux.die.net/man/8/iptables) are used. 
The [init-iptables container](https://quay.io/repository/fiware/init-iptables) sets to rules:

* ```iptables -t nat -A OUTPUT -m owner --uid-owner $ENVOY_USER_ID -j RETURN``` - return everything coming from envoy's user-id
* ```iptables -t nat -A OUTPUT -p tcp -j REDIRECT --to-ports $ENVOY_PORT``` - forwared every tcp traffic to envoy, using REDIRECT

In contrast to [istio][https://istio.io/], that uses a very similar approach, we only need those two, since we do not handle incoming traffic.