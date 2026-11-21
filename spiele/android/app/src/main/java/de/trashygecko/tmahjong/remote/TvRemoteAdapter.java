package de.trashygecko.tmahjong.remote;

import org.json.JSONObject;

/** Common asynchronous-protocol adapter contract. Methods are invoked off the UI thread. */
public interface TvRemoteAdapter {
    String id();
    JSONObject connect(JSONObject device);
    JSONObject send(JSONObject device,String command,JSONObject args);
    JSONObject listApps(JSONObject device);
    JSONObject launchApp(JSONObject device,String appId);
    default JSONObject pair(JSONObject device,String code){return unsupported("Für dieses Geräteprotokoll ist keine Kopplung implementiert.");}
    default JSONObject sendText(JSONObject device,String text){return unsupported("Texteingabe wird von diesem Geräteprotokoll nicht bestätigt.");}
    default JSONObject disconnect(JSONObject device){return new JSONObject();}
    default void forget(JSONObject device){}
    static JSONObject unsupported(String message){try{return new JSONObject().put("ok",false).put("connected",false).put("message",message);}catch(Exception e){return new JSONObject();}}
}