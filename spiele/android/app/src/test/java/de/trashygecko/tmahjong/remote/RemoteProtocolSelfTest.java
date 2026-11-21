package de.trashygecko.tmahjong.remote;

/** Dependency-free protocol assertions; run with Gradle task :app:remoteProtocolTest. */
public final class RemoteProtocolSelfTest {
    private static int checks;
    public static void main(String[] args) throws Exception {
        equal("Up", RokuEcpAdapter.keyForCommand("up", ""));
        equal("Down", RokuEcpAdapter.keyForCommand("down", ""));
        equal("Left", RokuEcpAdapter.keyForCommand("left", ""));
        equal("Right", RokuEcpAdapter.keyForCommand("right", ""));
        equal("Select", RokuEcpAdapter.keyForCommand("ok", ""));
        equal("Back", RokuEcpAdapter.keyForCommand("back", ""));
        equal("Home", RokuEcpAdapter.keyForCommand("home", ""));
        equal("Play", RokuEcpAdapter.keyForCommand("play", ""));
        equal("Play", RokuEcpAdapter.keyForCommand("pause", ""));
        equal("Rev", RokuEcpAdapter.keyForCommand("rewind", ""));
        equal("Fwd", RokuEcpAdapter.keyForCommand("fastForward", ""));
        equal("PowerOff", RokuEcpAdapter.keyForCommand("power", ""));
        equal("VolumeUp", RokuEcpAdapter.keyForCommand("volUp", ""));
        equal("VolumeDown", RokuEcpAdapter.keyForCommand("volDown", ""));
        equal("VolumeMute", RokuEcpAdapter.keyForCommand("mute", ""));
        equal("ChannelUp", RokuEcpAdapter.keyForCommand("channelUp", ""));
        equal("ChannelDown", RokuEcpAdapter.keyForCommand("channelDown", ""));
        equal("InputTuner", RokuEcpAdapter.keyForCommand("source", "TV"));
        equal("InputHDMI1", RokuEcpAdapter.keyForCommand("source", "HDMI1"));
        equal("InputHDMI4", RokuEcpAdapter.keyForCommand("source", "HDMI4"));
        equal("InputAV1", RokuEcpAdapter.keyForCommand("source", "AV"));
        equal(null, RokuEcpAdapter.keyForCommand("source", "SCART"));
        equal(null, RokuEcpAdapter.keyForCommand("menu", ""));
        equal(null, RokuEcpAdapter.keyForCommand("digit4", ""));
        equal(null, RokuEcpAdapter.keyForCommand("text", ""));
        String[] ssdp = RokuEcpAdapter.parseSsdpReply("HTTP/1.1 200 OK\r\nLOCATION: http://192.168.1.50:8060/\r\nUSN: uuid:roku:ecp:serial\r\nST: roku:ecp\r\n\r\n");
        equal("uuid:roku:ecp:serial", ssdp == null ? null : ssdp[0]);
        equal("http://192.168.1.50:8060/", ssdp == null ? null : ssdp[1]);
        equal(null, RokuEcpAdapter.parseSsdpReply("HTTP/1.1 200 OK\r\nLOCATION: http://192.168.1.40:8060/\r\nST: ssdp:all\r\n\r\n"));
        equal(null, RokuEcpAdapter.parseSsdpReply("HTTP/1.1 200 OK\r\nUSN: uuid:roku:ecp:missing-location\r\n"));
        checks++;
        if (!RokuEcpAdapter.isAllowedLocalAddress(java.net.InetAddress.getByName("192.168.1.2"))) throw new AssertionError("Private IPv4 should be allowed");
        checks++;
        if (RokuEcpAdapter.isAllowedLocalAddress(java.net.InetAddress.getByName("8.8.8.8"))) throw new AssertionError("Public IPv4 should be rejected");
        checks++;
        if (RokuEcpAdapter.isAllowedLocalAddress(java.net.InetAddress.getByName("127.0.0.1"))) throw new AssertionError("Loopback should be rejected");
        checks++;        if (!RemoteAdapterRegistry.unsupportedMessage("samsung").contains("verifiziert")) throw new AssertionError("Unsupported adapter must fail explicitly");
        checks++;
        equal(19, AndroidTvRemoteAdapter.keyCode("up", ""));
        equal(23, AndroidTvRemoteAdapter.keyCode("ok", ""));
        equal(26, AndroidTvRemoteAdapter.keyCode("power", ""));
        equal(8, AndroidTvRemoteAdapter.keyCode("digit1", ""));
        equal(243, AndroidTvRemoteAdapter.keyCode("source", "HDMI1"));
        equal(-1, AndroidTvRemoteAdapter.keyCode("source", "SCART"));
        String[] dial = LocalServiceDiscovery.parseReplyDetails("HTTP/1.1 200 OK\r\nLOCATION: http://192.168.1.60:8008/ssdp/device-desc.xml\r\nUSN: uuid:device\r\nST: urn:dial-multiscreen-org:service:dial:1\r\n\r\n");
        equal("dial", dial == null ? null : dial[0]);
        equal("192.168.1.60", dial == null ? null : dial[1]);
        equal(null, LocalServiceDiscovery.parseReplyDetails("HTTP/1.1 200 OK\r\nLOCATION: http://8.8.8.8/device.xml\r\nST: urn:dial-multiscreen-org:service:dial:1\r\n\r\n"));
        byte[] message = ProtoWire.outer(30, ProtoWire.stringField(1, "test"));
        java.util.List<ProtoWire.Field> outer = ProtoWire.parse(message);
        equal(2L, ProtoWire.first(outer, 1).integer);
        equal(200L, ProtoWire.first(outer, 2).integer);
        equal("test", ProtoWire.first(ProtoWire.parse(ProtoWire.first(outer, 30).data), 1).string());
        java.io.ByteArrayOutputStream frame = new java.io.ByteArrayOutputStream();
        ProtoWire.writeDelimited(frame, message);
        byte[] decoded = ProtoWire.readDelimited(new java.io.ByteArrayInputStream(frame.toByteArray()));
        equal(java.util.Arrays.toString(message), java.util.Arrays.toString(decoded));
        System.out.println("Remote-Protokoll-Selbsttest bestanden: " + checks + " Assertions");
    }
    private static void equal(Object expected, Object actual) {
        checks++;
        if (expected == null ? actual != null : !expected.equals(actual)) throw new AssertionError("Expected " + expected + ", got " + actual);
    }
}


