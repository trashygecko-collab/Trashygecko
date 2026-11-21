package de.trashygecko.tmahjong.remote;

import android.content.Context;
import android.net.wifi.WifiManager;
import org.json.JSONArray;
import org.json.JSONObject;
import java.net.DatagramPacket;
import java.net.DatagramSocket;
import java.net.InetAddress;
import java.net.InetSocketAddress;
import java.net.URL;
import java.nio.charset.StandardCharsets;
import java.util.ArrayList;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Locale;
import java.util.Map;

/** LAN discovery only. SSDP results are not fetched over HTTP and are never treated as proof of control support. */
public final class LocalServiceDiscovery {
    private static final String MULTICAST="239.255.255.250";
    private LocalServiceDiscovery(){}

    public static JSONObject discover(Context context,int timeoutMs){
        List<String>warnings=new ArrayList<>();Map<String,JSONObject> found=new LinkedHashMap<>();WifiManager wifi=(WifiManager)context.getApplicationContext().getSystemService(Context.WIFI_SERVICE);WifiManager.MulticastLock lock=null;
        try{
            if(wifi!=null){lock=wifi.createMulticastLock("TrashyRemote-SSDP");lock.setReferenceCounted(false);lock.acquire();}
            long deadline=System.currentTimeMillis()+Math.max(1500,Math.min(timeoutMs,4000));
            try(DatagramSocket socket=new DatagramSocket(null)){
                socket.setReuseAddress(true);socket.bind(new InetSocketAddress(0));socket.setSoTimeout(250);
                InetAddress target=InetAddress.getByName(MULTICAST);
                send(socket,target,"roku:ecp");send(socket,target,"urn:dial-multiscreen-org:service:dial:1");
                while(!Thread.currentThread().isInterrupted()&&System.currentTimeMillis()<deadline){
                    byte[] buffer=new byte[8192];DatagramPacket packet=new DatagramPacket(buffer,buffer.length);
                    try{socket.receive(packet);}catch(java.net.SocketTimeoutException timeout){continue;}
                    String response=new String(packet.getData(),packet.getOffset(),packet.getLength(),StandardCharsets.US_ASCII);
                    JSONObject device=parseReply(response);
                    if(device!=null){String host=device.optString("host");JSONObject previous=found.get(host);
                        if(previous==null||"roku".equals(device.optString("brand")))found.put(host,device);
                    }
                }
            }
        }catch(Exception error){if(Thread.currentThread().isInterrupted())warnings.add("SSDP-Suche wurde abgebrochen.");else warnings.add("SSDP-Suche fehlgeschlagen: "+safe(error));}
        finally{if(lock!=null&&lock.isHeld())lock.release();}
        if(found.isEmpty())warnings.add("Keine Roku-ECP- oder DIAL-SSDP-Antwort empfangen. Gastnetz-Isolation und WLAN-Multicast können die Suche verhindern.");
        JSONArray devices=new JSONArray();for(JSONObject item:found.values())devices.put(item);
        try{return new JSONObject().put("ok",true).put("devices",devices).put("warnings",new JSONArray(warnings));}catch(Exception e){return TvRemoteAdapter.unsupported("Suchergebnis konnte nicht erstellt werden.");}
    }

    private static void send(DatagramSocket socket,InetAddress target,String type)throws Exception{
        String request="M-SEARCH * HTTP/1.1\r\nHOST: "+MULTICAST+":1900\r\nMAN: \"ssdp:discover\"\r\nMX: 2\r\nST: "+type+"\r\n\r\n";
        byte[] bytes=request.getBytes(StandardCharsets.US_ASCII);socket.send(new DatagramPacket(bytes,bytes.length,target,1900));
    }

    /** Pure SSDP parser used by deterministic tests; no Android JSON runtime is required. */
    public static String[] parseReplyDetails(String response){
        if(response==null||!response.startsWith("HTTP/1."))return null;
        String location=header(response,"location"),usn=header(response,"usn"),st=header(response,"st");if(location.isEmpty())return null;
        String identity=(usn+" "+st).toLowerCase(Locale.ROOT);String brand,protocol,name;
        if(identity.contains("roku:ecp")){brand="roku";protocol="Roku ECP · SSDP";name="Roku-Gerät (nur erkannt)";}
        else if(identity.contains("dial-multiscreen-org:service:dial")){brand="dial";protocol="DIAL · SSDP";name="DIAL-Gerät (nur erkannt)";}
        else return null;
        try{
            URL url=new URL(location);if(!"http".equalsIgnoreCase(url.getProtocol()))return null;
            int port=url.getPort()<0?80:url.getPort();if(port<1||port>65535)return null;
            InetAddress address=InetAddress.getByName(url.getHost());if(!RokuEcpAdapter.isAllowedLocalAddress(address))return null;
            return new String[]{brand,address.getHostAddress(),Integer.toString(port),protocol,name};
        }catch(Exception invalid){return null;}
    }

    /** Deterministic parser used by bridge discovery; SSDP never fetches this URL. */
    public static JSONObject parseReply(String response){
        String[] fields=parseReplyDetails(response);if(fields==null)return null;
        try{return new JSONObject().put("id",fields[0]+":"+fields[1]).put("brand",fields[0])
                .put("manufacturer",fields[0].equals("roku")?"Roku":"DIAL-Gerät")
                .put("system",fields[0].equals("roku")?"Roku OS":"DIAL-fähiges Streaminggerät")
                .put("name",fields[4]).put("host",fields[1]).put("port",Integer.parseInt(fields[2]))
                .put("protocol",fields[3]).put("discoveryOnly",true).put("paired",false);
        }catch(Exception invalid){return null;}
    }
    private static String header(String response,String wanted){for(String line:response.split("\\r?\\n")){int colon=line.indexOf(':');if(colon>0&&line.substring(0,colon).trim().equalsIgnoreCase(wanted))return line.substring(colon+1).trim();}return "";}
    private static String safe(Exception e){return e.getMessage()==null||e.getMessage().isEmpty()?e.getClass().getSimpleName():e.getMessage();}
}