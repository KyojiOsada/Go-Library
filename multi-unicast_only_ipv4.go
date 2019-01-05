func SendMultiUnicast(addresses []string, message string) {
    // Start multi-unicast
    for _, address := range addresses {
        go func(_address string) {
            address_byte, err := net.ResolveUDPAddr("udp", _address)
            multi_unicast, err := net.DialUDP("udp", nil, address_byte)
            Error(err)

            defer multi_unicast.Close()
            fmt.Printf("Connected multi-unicast > %s\n", _address)

            // Multi-unicast message
            multi_unicast.Write([]byte(_message))
            fmt.Printf("Multi-unicast > %s as “%s”\n", _address, _message)
        }(address)
    }
}
