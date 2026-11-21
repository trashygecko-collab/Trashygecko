package de.trashygecko.tmahjong.remote;

import android.content.Context;
import org.json.JSONObject;
import java.net.InetAddress;
import java.net.URLDecoder;
import java.net.URLEncoder;
import java.util.Locale;

/** Roku discovery is allowed, but ECP control is disabled for mobile third-party clients per Roku policy. */
public final class RokuEcpAdapter implements TvRemoteAdapter {
    @SuppressWarnings("unused") private final Context appContext;
    public RokuEcpAdapter(Context context){appContext=context.getApplicationContext();}
    @Override public String id(){return "roku-ecp-discovery-only";}
    @Override public JSONObject connect(JSONObject device){return blocked();}
    @Override public JSONObject send(JSONObject device,String command,JSONObject args){return blocked();}
    @Override public JSONObject listApps(JSONObject device){return blocked();}
    @Override public JSONObject launchApp(JSONObject device,String appId){return blocked();}
    @Override public JSONObject pair(JSONObject device,String code){return blocked();}
    private static JSONObject blocked(){return TvRemoteAdapter.unsupported("Roku erlaubt laut aktueller ECP-Dokumentation keine ECP-Befehle von Drittanbieter-Mobilplattformen. Das Gerät wurde nur per SSDP erkannt; Fernsteuerung ist deaktiviert.");}

    /** Pure parsing helper retained for deterministic SSDP tests. */
    public static String[] parseSsdpReply(String response){
        if(response==null)return null;String location=header(response,"location"),usn=header(response,"usn"),service=header(response,"st");
        String identity=(usn+" "+service).toLowerCase(Locale.ROOT);
        return !location.isEmpty()&&identity.contains("roku:ecp")?new String[]{usn,location}:null;
    }
    public static boolean isAllowedLocalAddress(InetAddress a){return a!=null&&!a.isAnyLocalAddress()&&!a.isLoopbackAddress()&&!a.isMulticastAddress()&&(a.isSiteLocalAddress()||a.isLinkLocalAddress());}
    private static String header(String text,String wanted){for(String line:text.split("\\r?\\n")){int colon=line.indexOf(':');if(colon>0&&line.substring(0,colon).trim().equalsIgnoreCase(wanted))return line.substring(colon+1).trim();}return "";}
    public static String keyForCommand(String command,String source){if(command==null)return null;if(source==null)source="";switch(command){
        case "up":return "Up";case "down":return "Down";case "left":return "Left";case "right":return "Right";case "ok":return "Select";
        case "back":return "Back";case "home":return "Home";case "play":case "pause":return "Play";case "rewind":return "Rev";case "fastForward":return "Fwd";
        case "power":return "PowerOff";case "volUp":return "VolumeUp";case "volDown":return "VolumeDown";case "mute":return "VolumeMute";case "channelUp":return "ChannelUp";case "channelDown":return "ChannelDown";
        case "source":switch(source){case "TV":return "InputTuner";case "HDMI1":return "InputHDMI1";case "HDMI2":return "InputHDMI2";case "HDMI3":return "InputHDMI3";case "HDMI4":return "InputHDMI4";case "AV":return "InputAV1";default:return null;}default:return null;}}
    public static String encodePath(String value){try{return URLEncoder.encode(value,"UTF-8").replace("+","%20");}catch(Exception e){return "";}}
}