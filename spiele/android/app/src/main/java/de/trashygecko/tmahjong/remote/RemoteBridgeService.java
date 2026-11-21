package de.trashygecko.tmahjong.remote;

import android.content.Context;
import android.content.SharedPreferences;
import org.json.JSONArray;
import org.json.JSONObject;
import java.util.HashSet;
import java.util.Set;
import java.util.concurrent.ConcurrentHashMap;
import java.util.Map;

/** Validates async WebView calls and routes them to the native TV protocol adapters. */
public final class RemoteBridgeService {
    private final Context appContext;
    private final RemoteAdapterRegistry registry;
    private final AndroidTvRemoteAdapter androidTv;
    private final SharedPreferences saved;
    private volatile Thread activeDiscovery;
    private final Map<String,Set<String>> activeCapabilities=new ConcurrentHashMap<>();

    public RemoteBridgeService(Context context){appContext=context.getApplicationContext();androidTv=new AndroidTvRemoteAdapter(appContext);registry=new RemoteAdapterRegistry(new RokuEcpAdapter(appContext),androidTv);saved=appContext.getSharedPreferences("trashy_remote_devices",Context.MODE_PRIVATE);}

    public JSONObject execute(String method,JSONObject payload){if(payload==null)payload=new JSONObject();switch(method){
        case "discover":return discover(payload);case "stopDiscovery":return stopDiscovery();case "connect":return connect(payload.optJSONObject("device"));case "send":return send(payload);
        case "pair":return pair(payload);case "disconnect":return disconnect(payload);case "listApps":return withDevice(payload);case "launchApp":return launch(payload);
        case "sendText":return sendText(payload);case "saveDevice":return save(payload.optJSONObject("device"));case "removeDevice":return remove(payload);case "getStoredDevices":return stored();
        default:return fail("Unbekannte Bridge-Operation: "+safeName(method));}}

    public void shutdown(){androidTv.shutdown();activeCapabilities.clear();}

    private JSONObject discover(JSONObject payload){
        activeDiscovery=Thread.currentThread();try{
            int timeout=Math.max(2500,Math.min(payload.optInt("timeoutMs",8000),12000));
            JSONObject result=LocalServiceDiscovery.discover(appContext,Math.min(4000,timeout/2));
            JSONObject android=AndroidTvRemoteAdapter.discover(appContext,Math.min(4000,timeout/2));
            JSONArray devices=result.optJSONArray("devices"),warnings=result.optJSONArray("warnings");if(devices==null)devices=new JSONArray();if(warnings==null)warnings=new JSONArray();
            JSONArray androidDevices=android.optJSONArray("devices");if(androidDevices!=null)for(int i=0;i<androidDevices.length();i++){JSONObject item=androidDevices.optJSONObject(i);if(item!=null&&!containsHost(devices,item.optString("host")))devices.put(item);}
            JSONArray androidWarnings=android.optJSONArray("warnings");if(androidWarnings!=null)for(int i=0;i<androidWarnings.length();i++){String w=androidWarnings.optString(i);if(!w.isEmpty()&&!containsText(warnings,w))warnings.put(w);}
            return new JSONObject().put("ok",true).put("devices",devices).put("warnings",warnings);
        }catch(Exception error){return fail("Suche fehlgeschlagen: "+safe(error));}finally{activeDiscovery=null;}
    }
    private static boolean containsHost(JSONArray values,String host){for(int i=0;i<values.length();i++){JSONObject d=values.optJSONObject(i);if(d!=null&&host.equals(d.optString("host")))return true;}return false;}
    private static boolean containsText(JSONArray values,String text){for(int i=0;i<values.length();i++)if(text.equals(values.optString(i)))return true;return false;}
    private JSONObject stopDiscovery(){Thread t=activeDiscovery;if(t!=null)t.interrupt();return ok(t==null?"Es läuft gerade keine Suche.":"Die Netzwerksuche wird abgebrochen.");}

    private JSONObject connect(JSONObject device){if(device==null)return fail("Geräteangaben fehlen.");TvRemoteAdapter adapter=registry.forBrand(device.optString("brand"));if(adapter==null)return fail(RemoteAdapterRegistry.unsupportedMessage(device.optString("brand","unbekanntes System")));
        JSONObject result=adapter.connect(device);String key=deviceKey(device);if(result.optBoolean("connected")){JSONArray list=result.optJSONArray("capabilities");Set<String>values=new HashSet<>();if(list!=null)for(int i=0;i<list.length();i++)values.add(list.optString(i));activeCapabilities.put(key,values);}else activeCapabilities.remove(key);return result;}
    private JSONObject send(JSONObject payload){JSONObject device=payload.optJSONObject("device");if(device==null)return fail("Geräteangaben fehlen.");TvRemoteAdapter adapter=registry.forBrand(device.optString("brand"));if(adapter==null)return fail(RemoteAdapterRegistry.unsupportedMessage(device.optString("brand")));
        String command=payload.optString("command","");Set<String>caps=activeCapabilities.get(deviceKey(device));if(caps==null||!caps.contains(command))return fail("Verbinde das Gerät erneut; diese Taste wurde nicht als verfügbar bestätigt.");return adapter.send(device,command,payload.optJSONObject("args"));}
    private JSONObject pair(JSONObject payload){JSONObject device=payload.optJSONObject("device");if(device==null)return fail("Geräteangaben fehlen.");TvRemoteAdapter adapter=registry.forBrand(device.optString("brand"));if(adapter==null)return fail(RemoteAdapterRegistry.unsupportedMessage(device.optString("brand")));return adapter.pair(device,payload.optString("code",""));}
    private JSONObject sendText(JSONObject payload){JSONObject device=payload.optJSONObject("device");if(device==null)return fail("Geräteangaben fehlen.");Set<String>caps=activeCapabilities.get(deviceKey(device));if(caps==null||!caps.contains("text"))return fail("Dieses Gerät hat Texteingabe nicht bestätigt.");TvRemoteAdapter adapter=registry.forBrand(device.optString("brand"));return adapter==null?fail(RemoteAdapterRegistry.unsupportedMessage(device.optString("brand"))):adapter.sendText(device,payload.optString("text",""));}
    private JSONObject disconnect(JSONObject payload){JSONObject device=payload.optJSONObject("device");if(device==null)return fail("Geräteangaben fehlen.");activeCapabilities.remove(deviceKey(device));TvRemoteAdapter adapter=registry.forBrand(device.optString("brand"));return adapter==null?ok("Verbindung getrennt."):adapter.disconnect(device);}
    private JSONObject withDevice(JSONObject payload){JSONObject device=payload.optJSONObject("device");if(device==null)return fail("Geräteangaben fehlen.");TvRemoteAdapter adapter=registry.forBrand(device.optString("brand"));if(adapter==null)return fail(RemoteAdapterRegistry.unsupportedMessage(device.optString("brand")));Set<String>caps=activeCapabilities.get(deviceKey(device));if(caps==null||!caps.contains("apps"))return fail("Das Gerät bestätigt keine App-Liste.");return adapter.listApps(device);}
    private JSONObject launch(JSONObject payload){JSONObject device=payload.optJSONObject("device");if(device==null)return fail("Geräteangaben fehlen.");TvRemoteAdapter adapter=registry.forBrand(device.optString("brand"));if(adapter==null)return fail(RemoteAdapterRegistry.unsupportedMessage(device.optString("brand")));Set<String>caps=activeCapabilities.get(deviceKey(device));if(caps==null||!caps.contains("apps"))return fail("App-Start ist nicht über eine bestätigte App-Liste verfügbar.");return adapter.launchApp(device,payload.optString("appId",""));}

    private synchronized JSONObject save(JSONObject device){if(device==null)return fail("Geräteangaben fehlen.");String id=device.optString("id","").trim(),host=device.optString("host","").trim();if(id.isEmpty()||id.length()>180||host.length()>253)return fail("Geräte-ID oder Adresse ist ungültig.");
        try{JSONObject clean=new JSONObject().put("id",id).put("host",host).put("brand",bounded(device.optString("brand"),40)).put("name",bounded(device.optString("name"),40)).put("model",bounded(device.optString("model"),80)).put("room",bounded(device.optString("room"),40)).put("manufacturer",bounded(device.optString("manufacturer"),60)).put("system",bounded(device.optString("system"),80)).put("protocol",bounded(device.optString("protocol"),100));JSONArray all=readSaved();boolean found=false;for(int i=0;i<all.length();i++)if(id.equals(all.getJSONObject(i).optString("id"))){all.put(i,clean);found=true;break;}if(!found)all.put(clean);saved.edit().putString("devices",all.toString()).apply();return ok("Gerätemetadaten lokal gespeichert.");}
        catch(Exception e){return fail("Gerät konnte nicht gespeichert werden.");}}
    private synchronized JSONObject remove(JSONObject payload){JSONObject device=payload.optJSONObject("device");String id=payload.optString("id",device==null?"":device.optString("id","")).trim();if(id.isEmpty())return fail("Geräte-ID fehlt.");if(device!=null){TvRemoteAdapter adapter=registry.forBrand(device.optString("brand"));if(adapter!=null)adapter.forget(device);}activeCapabilities.remove(id);
        try{JSONArray input=readSaved(),out=new JSONArray();for(int i=0;i<input.length();i++)if(!id.equals(input.getJSONObject(i).optString("id")))out.put(input.getJSONObject(i));saved.edit().putString("devices",out.toString()).apply();return ok("Gerät und lokale Kopplungsreferenz entfernt.");}catch(Exception e){return fail("Gerät konnte nicht entfernt werden.");}}
    private synchronized JSONObject stored(){try{return new JSONObject().put("ok",true).put("devices",readSaved());}catch(Exception e){return fail("Gespeicherte Geräte konnten nicht geladen werden.");}}
    private JSONArray readSaved()throws Exception{JSONArray source=new JSONArray(saved.getString("devices","[]")),clean=new JSONArray();for(int i=0;i<source.length();i++){JSONObject r=source.optJSONObject(i);if(r==null)continue;clean.put(new JSONObject().put("id",bounded(r.optString("id"),180)).put("host",bounded(r.optString("host"),253)).put("brand",bounded(r.optString("brand"),40)).put("name",bounded(r.optString("name"),40)).put("model",bounded(r.optString("model"),80)).put("room",bounded(r.optString("room"),40)).put("manufacturer",bounded(r.optString("manufacturer"),60)).put("system",bounded(r.optString("system"),80)).put("protocol",bounded(r.optString("protocol"),100)));}return clean;}
    private static String deviceKey(JSONObject d){if(d==null)return "";String id=d.optString("id","").trim();return id.isEmpty()?d.optString("host","").trim():id;}
    private static String bounded(String value,int limit){String t=value==null?"":value.trim();return t.substring(0,Math.min(t.length(),limit));}
    private static String safeName(String value){return bounded(value,40).replaceAll("[^A-Za-z0-9_-]","?");}
    private static String safe(Exception e){return e.getMessage()==null||e.getMessage().trim().isEmpty()?e.getClass().getSimpleName():e.getMessage();}
    private static JSONObject ok(String message){try{return new JSONObject().put("ok",true).put("message",message);}catch(Exception e){return new JSONObject();}}
    private static JSONObject fail(String message){try{return new JSONObject().put("ok",false).put("connected",false).put("message",message);}catch(Exception e){return new JSONObject();}}
}