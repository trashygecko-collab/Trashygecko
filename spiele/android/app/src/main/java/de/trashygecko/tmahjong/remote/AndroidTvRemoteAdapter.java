package de.trashygecko.tmahjong.remote;

import android.content.Context;
import android.content.SharedPreferences;
import android.os.Build;
import android.security.KeyPairGeneratorSpec;
import android.security.keystore.KeyGenParameterSpec;
import android.security.keystore.KeyProperties;
import android.util.Base64;
import org.json.JSONArray;
import org.json.JSONObject;
import java.io.IOException;
import java.math.BigInteger;
import java.net.Inet4Address;
import java.net.InetAddress;
import java.net.InetSocketAddress;
import java.net.Socket;
import java.nio.charset.StandardCharsets;
import java.security.KeyPairGenerator;
import java.security.KeyStore;
import java.security.MessageDigest;
import java.security.SecureRandom;
import java.security.cert.X509Certificate;
import java.security.interfaces.RSAPublicKey;
import java.util.ArrayList;
import java.util.Calendar;
import java.util.Collections;
import java.util.HashSet;
import java.util.List;
import java.util.Locale;
import java.util.Set;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.CountDownLatch;
import java.util.concurrent.ConcurrentLinkedQueue;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicBoolean;
import java.util.concurrent.atomic.AtomicReference;
import javax.net.ssl.KeyManagerFactory;
import javax.net.ssl.SSLContext;
import javax.net.ssl.SSLSocket;
import javax.net.ssl.TrustManager;
import javax.net.ssl.X509TrustManager;
import javax.security.auth.x500.X500Principal;

/** Android TV Remote v2 (TLS + TV PIN). Private client identity is kept in AndroidKeyStore. */
public final class AndroidTvRemoteAdapter implements TvRemoteAdapter {
    private static final String ALIAS="trashyremote-atvremote-v2", SERVICE="atvremote";
    private static final int PAIR_PORT=6467, REMOTE_PORT=6466;
    private static final int PING=1, KEY=2, IME=4, POWER=32, VOLUME=64, APP_LINK=512;
    private static final int REQUESTED=PING|KEY|IME|POWER|VOLUME|APP_LINK;
    private static final long PAIR_MS=90_000L;
    private final Context context;
    private final SharedPreferences pins;
    private final ConcurrentHashMap<String,Pending> pending=new ConcurrentHashMap<>();
    private final ConcurrentHashMap<String,Session> sessions=new ConcurrentHashMap<>();

    public AndroidTvRemoteAdapter(Context c){context=c.getApplicationContext();pins=context.getSharedPreferences("trashy_remote_atv_trust",Context.MODE_PRIVATE);}
    @Override public String id(){return "android-tv-remote-v2";}

    @Override public JSONObject connect(JSONObject device){
        String host=host(device), key=deviceKey(device);
        if(key.isEmpty())return fail("Geräte-ID fehlt. Speichere das Android-TV-Gerät zuerst.");
        if(!isPrivateHost(host))return fail("Android TV benötigt eine private lokale IPv4-Adresse.");
        Session old=sessions.get(key); if(old!=null&&old.isOpen())return result(old,host,"Die verschlüsselte Android-TV-Verbindung besteht bereits.");
        closeSession(key);
        String pin=pins.getString(pinKey(key),null);
        if(pin==null){try{startPairing(key,host);return pairing("Auf dem Fernseher wird jetzt ein sechsstelliger HEX-Code angezeigt. Gib genau diesen Code ein; er wird nicht gespeichert.");}catch(Exception e){return fail(error("Android-TV-Kopplung konnte nicht gestartet werden",e));}}
        try{Session s=openRemote(key,host,pin);sessions.put(key,s);s.startReader();return result(s,host,"Android TV ist über TLS verbunden. Gerätefähigkeiten wurden ausgehandelt.");}
        catch(Exception e){closeSession(key);String m=e.getMessage();if(m!=null&&m.contains("TV-Zertifikat"))return fail("Die gepinnte TV-Identität stimmt nicht mehr. Entferne dieses Gerät und kopple es nach Prüfung neu.");return fail(error("Android-TV-Verbindung fehlgeschlagen",e));}
    }

    @Override public JSONObject pair(JSONObject device,String code){
        String key=deviceKey(device); if(key.isEmpty())return fail("Geräte-ID fehlt.");
        Pending p=pending.get(key);
        if(p==null||System.currentTimeMillis()>p.expires){closePending(key);String h=host(device);if(!isPrivateHost(h))return fail("Android TV benötigt eine private lokale IPv4-Adresse.");try{startPairing(key,h);return pairing("Der Kopplungscode wurde neu angefordert. Lies den aktuellen sechsstelligen HEX-Code am TV ab und gib ihn erneut ein.");}catch(Exception e){return fail(error("Android-TV-Kopplung konnte nicht gestartet werden",e));}}
        String pin=code==null?"":code.trim().replace(" ","").toUpperCase(Locale.ROOT);
        if(!pin.matches("[0-9A-F]{6}"))return fail("Der Android-TV-Code muss genau sechs Zeichen (0–9, A–F) enthalten.");
        try{
            byte[] digest=pairingDigest(p.client,p.server,pin);
            if((digest[0]&255)!=Integer.parseInt(pin.substring(0,2),16))return fail("Der Code passt nicht zu diesem Fernseher. Prüfe die sechs Zeichen.");
            ProtoWire.writeDelimited(p.socket.getOutputStream(),ProtoWire.outer(40,ProtoWire.bytesField(1,digest)));
            byte[] ack=ProtoWire.readDelimited(p.socket.getInputStream());checkStatus(ack);
            if(ProtoWire.first(ProtoWire.parse(ack),41)==null)throw new IOException("Der Fernseher hat die Kopplung nicht bestätigt.");
            pins.edit().putString(pinKey(key),fingerprint(p.server)).apply();pending.remove(key,p);close(p.socket);
            return new JSONObject().put("ok",true).put("paired",true).put("message","TV hat die Kopplung bestätigt. Der private Schlüssel bleibt im Android Keystore; das TV-Zertifikat ist lokal angeheftet.");
        }catch(Exception e){return fail(error("Android-TV-Kopplung nicht bestätigt",e));}
    }

    @Override public JSONObject send(JSONObject device,String command,JSONObject args){
        String key=deviceKey(device);Session s=sessions.get(key);if(s==null||!s.isOpen())return lost("Android TV ist getrennt. Verbinde das Gerät erneut.");
        int code=keyCode(command,args==null?"":args.optString("source",""));if(code<0)return fail("Dieser Android-TV-Befehl ist nicht unterstützt.");
        if(!s.caps.contains(command))return fail("Das Gerät hat diese Taste nicht als verfügbar bestätigt.");
        try{byte[] body=ProtoWire.join(ProtoWire.varintField(1,code),ProtoWire.varintField(2,3));s.send(ProtoWire.remote(10,body));return ok("Tastenereignis über TLS übertragen.");}
        catch(Exception e){closeSession(key);return lost(error("Android-TV-Befehl fehlgeschlagen",e));}
    }

    @Override public JSONObject sendText(JSONObject device,String text){
        String key=deviceKey(device);Session s=sessions.get(key);if(s==null||!s.isOpen())return lost("Android TV ist getrennt. Verbinde das Gerät erneut.");
        if(!s.caps.contains("text"))return fail("Das Android-TV-Gerät hat keine Texteingabe bestätigt.");
        if(text==null||text.isEmpty()||text.length()>256)return fail("Der Text muss 1 bis 256 Zeichen enthalten.");
        try{
            byte[] field=ProtoWire.join(ProtoWire.varintField(1,text.length()),ProtoWire.varintField(2,text.length()),ProtoWire.stringField(3,text));
            byte[] edit=ProtoWire.join(ProtoWire.varintField(1,1),ProtoWire.bytesField(2,field));
            byte[] batch=ProtoWire.join(ProtoWire.varintField(1,s.imeCounter),ProtoWire.varintField(2,s.imeFieldCounter),ProtoWire.bytesField(3,edit));
            s.send(ProtoWire.remote(21,batch));return ok("Text über die bestätigte Android-TV-Eingabefunktion übertragen.");
        }catch(Exception e){closeSession(key);return lost(error("Texteingabe fehlgeschlagen",e));}
    }

    @Override public JSONObject listApps(JSONObject device){return fail("Android TV Remote v2 meldet keine installierte App-Liste.");}
    @Override public JSONObject launchApp(JSONObject device,String appId){return fail("App-Start benötigt einen bestätigten Deep Link; eine App-Liste oder generische Paket-ID-Steuerung ist nicht verfügbar.");}
    @Override public JSONObject disconnect(JSONObject device){String key=deviceKey(device);closeSession(key);closePending(key);return disconnected("Android-TV-Verbindung geschlossen.");}
    @Override public void forget(JSONObject device){String key=deviceKey(device);closeSession(key);closePending(key);if(!key.isEmpty())pins.edit().remove(pinKey(key)).apply();}
    public void shutdown(){for(String key:new ArrayList<>(sessions.keySet()))closeSession(key);for(String key:new ArrayList<>(pending.keySet()))closePending(key);}
    private void startPairing(String key,String host)throws Exception{
        closePending(key);AtomicReference<X509Certificate> captured=new AtomicReference<>();SSLSocket socket=socket(host,PAIR_PORT,null,true,captured);boolean keep=false;
        try{
            ProtoWire.writeDelimited(socket.getOutputStream(),ProtoWire.outer(10,ProtoWire.join(ProtoWire.stringField(1,SERVICE),ProtoWire.stringField(2,"TrashyGecko Remote"))));
            byte[] ack=ProtoWire.readDelimited(socket.getInputStream());checkStatus(ack);if(ProtoWire.first(ProtoWire.parse(ack),11)==null)throw new IOException("Android TV hat die Pairing-Anfrage nicht bestätigt.");
            byte[] enc=ProtoWire.join(ProtoWire.varintField(1,3),ProtoWire.varintField(2,6));
            ProtoWire.writeDelimited(socket.getOutputStream(),ProtoWire.outer(20,ProtoWire.join(ProtoWire.bytesField(1,enc),ProtoWire.varintField(3,1))));
            byte[] opts=ProtoWire.readDelimited(socket.getInputStream());checkStatus(opts);if(ProtoWire.first(ProtoWire.parse(opts),20)==null)throw new IOException("TV bietet keine HEX-PIN-Kopplung an.");
            ProtoWire.writeDelimited(socket.getOutputStream(),ProtoWire.outer(30,ProtoWire.join(ProtoWire.bytesField(1,enc),ProtoWire.varintField(2,1))));
            byte[] configured=ProtoWire.readDelimited(socket.getInputStream());checkStatus(configured);if(ProtoWire.first(ProtoWire.parse(configured),31)==null)throw new IOException("Android TV hat die PIN-Kopplung nicht gestartet.");
            X509Certificate server=captured.get();if(server==null||!(server.getPublicKey() instanceof RSAPublicKey))throw new IOException("TV-Zertifikat fehlt oder ist kein RSA-Zertifikat.");
            KeyStore ks=clientStore();X509Certificate client=(X509Certificate)ks.getCertificate(ALIAS);if(client==null)throw new IOException("Clientzertifikat konnte nicht geladen werden.");
            pending.put(key,new Pending(socket,client,server,System.currentTimeMillis()+PAIR_MS));keep=true;
        }finally{if(!keep)close(socket);}
    }

    private Session openRemote(String key,String host,String pin)throws Exception{
        SSLSocket socket=socket(host,REMOTE_PORT,pin,false,new AtomicReference<X509Certificate>());boolean keep=false;
        try{
            Session s=new Session(key,host,socket);socket.setSoTimeout(4500);long deadline=System.currentTimeMillis()+9000;boolean configured=false,started=false;
            while(!started&&System.currentTimeMillis()<deadline){
                List<ProtoWire.Field> fields=ProtoWire.parse(ProtoWire.readDelimited(socket.getInputStream()));
                for(ProtoWire.Field f:fields){
                    if(f.number==1&&f.data!=null){List<ProtoWire.Field> cfg=ProtoWire.parse(f.data);ProtoWire.Field mask=ProtoWire.first(cfg,1);s.features=mask==null?0:(int)mask.integer&REQUESTED;ProtoWire.Field info=ProtoWire.first(cfg,2);if(info!=null&&info.data!=null)readDeviceInfo(s,info.data);if((s.features&KEY)==0)throw new IOException("Android-TV-Dienst bietet keine Tastensteuerung an.");ProtoWire.writeDelimited(socket.getOutputStream(),encodeConfigure(s.features));configured=true;}
                    else if(f.number==2&&f.data!=null)ProtoWire.writeDelimited(socket.getOutputStream(),ProtoWire.remote(2,ProtoWire.varintField(1,s.features)));
                    else if(f.number==8&&f.data!=null){ProtoWire.Field ping=ProtoWire.first(ProtoWire.parse(f.data),1);ProtoWire.writeDelimited(socket.getOutputStream(),ProtoWire.remote(9,ProtoWire.varintField(1,ping==null?0:ping.integer)));}
                    else if(f.number==40&&f.data!=null)started=true;
                    else if(f.number==3)throw new IOException("Android-TV-Dienst hat die Remote-Sitzung abgelehnt.");
                }
            }
            if(!configured||!started)throw new IOException("Android-TV-Remote-Dienst hat den Start nicht bestätigt.");
            s.buildCaps();socket.setSoTimeout(20000);keep=true;return s;
        }finally{if(!keep)close(socket);}
    }

    private static byte[] encodeConfigure(int features){
        byte[] info=ProtoWire.join(ProtoWire.varintField(3,1),ProtoWire.stringField(4,"1"),ProtoWire.stringField(5,"de.trashygecko.remote"),ProtoWire.stringField(6,"1.12"));
        return ProtoWire.remote(1,ProtoWire.join(ProtoWire.varintField(1,features),ProtoWire.bytesField(2,info)));
    }

    private SSLSocket socket(String host,int port,String expected,boolean allowUnpinned,AtomicReference<X509Certificate> captured)throws Exception{
        InetAddress target=resolvePrivate(host);ensureIdentity();KeyStore store=clientStore();
        KeyManagerFactory kmf=KeyManagerFactory.getInstance(KeyManagerFactory.getDefaultAlgorithm());kmf.init(store,null);
        SSLContext ssl=SSLContext.getInstance("TLS");ssl.init(kmf.getKeyManagers(),new TrustManager[]{new ServerTrust(expected,allowUnpinned,captured)},new SecureRandom());
        Socket raw=new Socket();raw.connect(new InetSocketAddress(target,port),3500);raw.setSoTimeout(5000);
        try{SSLSocket s=(SSLSocket)ssl.getSocketFactory().createSocket(raw,host,port,true);s.setEnabledProtocols(new String[]{"TLSv1.2"});s.setSoTimeout(5000);s.startHandshake();return s;}catch(Exception e){close(raw);throw e;}
    }

    private synchronized void ensureIdentity()throws Exception{
        KeyStore store=KeyStore.getInstance("AndroidKeyStore");store.load(null);if(store.containsAlias(ALIAS))return;
        Calendar begin=Calendar.getInstance(),end=Calendar.getInstance();end.add(Calendar.YEAR,20);KeyPairGenerator gen=KeyPairGenerator.getInstance("RSA","AndroidKeyStore");
        if(Build.VERSION.SDK_INT>=23){
            KeyGenParameterSpec spec=new KeyGenParameterSpec.Builder(ALIAS,KeyProperties.PURPOSE_SIGN|KeyProperties.PURPOSE_DECRYPT).setKeySize(2048)
                    .setDigests(KeyProperties.DIGEST_SHA256,KeyProperties.DIGEST_SHA384,KeyProperties.DIGEST_SHA512).setSignaturePaddings(KeyProperties.SIGNATURE_PADDING_RSA_PKCS1)
                    .setEncryptionPaddings(KeyProperties.ENCRYPTION_PADDING_RSA_PKCS1).setCertificateSubject(new X500Principal("CN=TrashyGecko Remote"))
                    .setCertificateSerialNumber(BigInteger.valueOf(System.currentTimeMillis())).setCertificateNotBefore(begin.getTime()).setCertificateNotAfter(end.getTime()).build();gen.initialize(spec);
        }else{
            KeyPairGeneratorSpec spec=new KeyPairGeneratorSpec.Builder(context).setAlias(ALIAS).setSubject(new X500Principal("CN=TrashyGecko Remote"))
                    .setSerialNumber(BigInteger.valueOf(System.currentTimeMillis())).setStartDate(begin.getTime()).setEndDate(end.getTime()).setKeySize(2048).build();gen.initialize(spec);
        }
        gen.generateKeyPair();
    }
    private static KeyStore clientStore()throws Exception{KeyStore s=KeyStore.getInstance("AndroidKeyStore");s.load(null);if(!s.containsAlias(ALIAS))throw new IOException("AndroidKeyStore-Identität fehlt.");return s;}

    private static void checkStatus(byte[] raw)throws IOException{ProtoWire.Field f=ProtoWire.first(ProtoWire.parse(raw),2);if(f==null||f.integer!=200)throw new IOException("Android-TV-Dienst hat die Protokollanfrage abgelehnt.");}
    private static void readDeviceInfo(Session s,byte[] raw)throws IOException{List<ProtoWire.Field> f=ProtoWire.parse(raw);ProtoWire.Field model=ProtoWire.first(f,1),vendor=ProtoWire.first(f,2),version=ProtoWire.first(f,6);s.model=model==null?"":model.string();s.manufacturer=vendor==null?"Android TV":vendor.string();s.version=version==null?"":version.string();}
    private static byte[] pairingDigest(X509Certificate client,X509Certificate server,String pin)throws Exception{
        RSAPublicKey c=(RSAPublicKey)client.getPublicKey(),s=(RSAPublicKey)server.getPublicKey();MessageDigest d=MessageDigest.getInstance("SHA-256");d.update(unsigned(c.getModulus()));d.update(unsigned(c.getPublicExponent()));d.update(unsigned(s.getModulus()));d.update(unsigned(s.getPublicExponent()));d.update(hex(pin.substring(2)));return d.digest();
    }
    private static byte[] unsigned(BigInteger n){byte[] b=n.toByteArray();return b.length>1&&b[0]==0?java.util.Arrays.copyOfRange(b,1,b.length):b;}
    private static byte[] hex(String s){byte[] b=new byte[s.length()/2];for(int i=0;i<b.length;i++)b[i]=(byte)Integer.parseInt(s.substring(i*2,i*2+2),16);return b;}
    private static String fingerprint(X509Certificate c)throws Exception{byte[] b=MessageDigest.getInstance("SHA-256").digest(c.getEncoded());StringBuilder s=new StringBuilder();for(byte v:b)s.append(String.format(Locale.ROOT,"%02x",v&255));return s.toString();}
    private static InetAddress resolvePrivate(String host)throws Exception{
        if(host==null||host.trim().isEmpty()||host.length()>253||host.matches(".*[\\s/\\\\?#@].*"))throw new IOException("Ungültige Geräteadresse.");
        InetAddress selected=null;for(InetAddress a:InetAddress.getAllByName(host.trim())){if(!(a instanceof Inet4Address)||a.isAnyLocalAddress()||a.isLoopbackAddress()||a.isMulticastAddress()||!(a.isSiteLocalAddress()||a.isLinkLocalAddress()))throw new IOException("Geräteadresse muss auf ein privates lokales IPv4-Ziel zeigen.");if(selected==null)selected=a;}if(selected==null)throw new IOException("Keine private IPv4-Adresse gefunden.");return selected;
    }
    public static boolean isPrivateHost(String host){try{resolvePrivate(host);return true;}catch(Exception e){return false;}}
    private static String host(JSONObject d){return d==null?"":d.optString("host",d.optString("ip","")).trim();}
    private static String deviceKey(JSONObject d){String k=d==null?"":d.optString("id","").trim();return k.isEmpty()?host(d):k;}
    private static String pinKey(String key){try{return "pin-"+Base64.encodeToString(MessageDigest.getInstance("SHA-256").digest(key.getBytes(StandardCharsets.UTF_8)),Base64.NO_WRAP|Base64.URL_SAFE);}catch(Exception e){return "pin-unknown";}}
    /** Discover only devices advertising the Android TV Remote v2 DNS-SD service. */
    public static JSONObject discover(Context context,int timeoutMs){
        JSONArray devices=new JSONArray();List<String>warnings=new ArrayList<>();android.net.nsd.NsdManager manager=(android.net.nsd.NsdManager)context.getApplicationContext().getSystemService(Context.NSD_SERVICE);
        if(manager==null)return discovery(devices,new String[]{"Android hat keinen mDNS/NSD-Dienst bereitgestellt."});
        final String type="_androidtvremote2._tcp.";final Object lock=new Object();final Set<String>seen=Collections.synchronizedSet(new HashSet<String>());final ConcurrentLinkedQueue<android.net.nsd.NsdServiceInfo>queue=new ConcurrentLinkedQueue<>();final AtomicBoolean resolving=new AtomicBoolean(false),stopped=new AtomicBoolean(false);
        final android.net.nsd.NsdManager.ResolveListener[] resolver=new android.net.nsd.NsdManager.ResolveListener[1];final Runnable[] next=new Runnable[1];
        next[0]=()->{if(stopped.get())return;android.net.nsd.NsdServiceInfo service=queue.poll();if(service==null){resolving.set(false);return;}try{manager.resolveService(service,resolver[0]);}catch(RuntimeException e){resolving.set(false);next[0].run();}};
        resolver[0]=new android.net.nsd.NsdManager.ResolveListener(){
            @Override public void onResolveFailed(android.net.nsd.NsdServiceInfo serviceInfo,int errorCode){resolving.set(false);next[0].run();}
            @Override public void onServiceResolved(android.net.nsd.NsdServiceInfo serviceInfo){
                InetAddress address=serviceInfo.getHost();String name=serviceInfo.getServiceName();
                if(address instanceof Inet4Address&&local(address)&&seen.add(name))synchronized(lock){try{devices.put(new JSONObject().put("id","androidtv:"+name).put("brand","android")
                        .put("manufacturer","").put("system","Android TV / Google TV").put("protocol","Android TV Remote v2 über TLS").put("name",name)
                        .put("host",address.getHostAddress()).put("port",serviceInfo.getPort()));}catch(Exception ignored){}}
                resolving.set(false);next[0].run();
            }
        };
        final android.net.nsd.NsdManager.DiscoveryListener listener=new android.net.nsd.NsdManager.DiscoveryListener(){
            @Override public void onDiscoveryStarted(String serviceType){}
            @Override public void onStartDiscoveryFailed(String serviceType,int errorCode){warnings.add("mDNS-Suche für Android TV Remote v2 konnte nicht gestartet werden.");}
            @Override public void onStopDiscoveryFailed(String serviceType,int errorCode){}
            @Override public void onDiscoveryStopped(String serviceType){}
            @Override public void onServiceLost(android.net.nsd.NsdServiceInfo serviceInfo){}
            @Override public void onServiceFound(android.net.nsd.NsdServiceInfo serviceInfo){if(serviceInfo.getServiceType()!=null&&serviceInfo.getServiceType().toLowerCase(Locale.ROOT).startsWith(type)){queue.add(serviceInfo);if(resolving.compareAndSet(false,true))next[0].run();}}
        };
        try{manager.discoverServices(type,android.net.nsd.NsdManager.PROTOCOL_DNS_SD,listener);Thread.sleep(Math.max(1000,Math.min(timeoutMs,4000)));}
        catch(InterruptedException e){Thread.currentThread().interrupt();warnings.add("mDNS-Suche wurde abgebrochen.");}
        catch(RuntimeException e){warnings.add("mDNS-Suche fehlgeschlagen: "+safe(e));}
        finally{stopped.set(true);try{manager.stopServiceDiscovery(listener);}catch(RuntimeException ignored){}}
        if(devices.length()==0)warnings.add("Kein Android-TV-Remote-v2-Dienst per mDNS gefunden. Modell, Firmware, TV-Freigabe und WLAN-Isolation können die Erkennung beeinflussen.");
        return discovery(devices,warnings.toArray(new String[0]));
    }

    public static int keyCode(String command,String source){
        if(command==null)return -1;
        switch(command){
            case "up":return 19;case "down":return 20;case "left":return 21;case "right":return 22;case "ok":return 23;
            case "back":return 4;case "home":return 3;case "menu":return 82;case "settings":return 176;case "power":return 26;
            case "volUp":return 24;case "volDown":return 25;case "mute":return 164;case "play":return 126;case "pause":return 127;case "playPause":return 85;
            case "rewind":return 89;case "fastForward":return 90;case "channelUp":return 166;case "channelDown":return 167;case "info":return 165;case "guide":return 172;
            case "digit0":return 7;case "digit1":return 8;case "digit2":return 9;case "digit3":return 10;case "digit4":return 11;
            case "digit5":return 12;case "digit6":return 13;case "digit7":return 14;case "digit8":return 15;case "digit9":return 16;
            case "source":if("HDMI1".equals(source))return 243;if("HDMI2".equals(source))return 244;if("HDMI3".equals(source))return 245;if("HDMI4".equals(source))return 246;if("TV".equals(source))return 170;return -1;
            default:return -1;
        }
    }

    private static boolean local(InetAddress a){return a!=null&&!a.isAnyLocalAddress()&&!a.isLoopbackAddress()&&!a.isMulticastAddress()&&(a.isSiteLocalAddress()||a.isLinkLocalAddress());}
    private static JSONObject discovery(JSONArray devices,String[]warnings){try{JSONArray w=new JSONArray();for(String s:warnings)if(s!=null&&!s.isEmpty())w.put(s);return new JSONObject().put("ok",true).put("devices",devices).put("warnings",w);}catch(Exception e){return fail("Suchergebnis konnte nicht aufgebaut werden.");}}
    private static JSONObject result(Session s,String host,String message){try{JSONArray c=new JSONArray();for(String v:s.caps)c.put(v);return new JSONObject().put("ok",true).put("connected",true).put("paired",true).put("host",host).put("system","Android TV / Google TV").put("protocol","Android TV Remote v2 über TLS").put("manufacturer",s.manufacturer).put("model",s.model).put("deviceVersion",s.version).put("capabilities",c).put("message",message);}catch(Exception e){return fail("Verbindungsstatus konnte nicht aufgebaut werden.");}}
    private static JSONObject pairing(String message){try{return new JSONObject().put("ok",false).put("connected",false).put("requiresPairing",true).put("pairingRequired",true).put("state","pairing").put("message",message);}catch(Exception e){return new JSONObject();}}
    private static JSONObject fail(String message){try{return new JSONObject().put("ok",false).put("connected",false).put("message",message);}catch(Exception e){return new JSONObject();}}
    private static JSONObject lost(String message){try{return new JSONObject().put("ok",false).put("connected",false).put("connectionLost",true).put("message",message);}catch(Exception e){return new JSONObject();}}
    private static JSONObject ok(String message){try{return new JSONObject().put("ok",true).put("message",message);}catch(Exception e){return new JSONObject();}}
    private static JSONObject disconnected(String message){try{return new JSONObject().put("ok",true).put("connected",false).put("message",message);}catch(Exception e){return new JSONObject();}}
    private static String error(String prefix,Exception e){
        String m=safe(e),l=m.toLowerCase(Locale.ROOT);
        if(l.contains("econnrefused")||l.contains("connection refused"))return prefix+": Der Fernseher nimmt keine Verbindung an (Port 6467 geschlossen). Ist er eingeschaltet und im selben WLAN?";
        if(l.contains("etimedout")||l.contains("timed out")||l.contains("timeout"))return prefix+": Zeitüberschreitung. Prüfe WLAN und ob der Fernseher erreichbar ist.";
        if(l.contains("enetunreach")||l.contains("network is unreachable")||l.contains("no route"))return prefix+": Das Gerät ist aus diesem Netz nicht erreichbar.";
        if(l.contains("econnreset")||l.contains("connection reset"))return prefix+": Der Fernseher hat die Verbindung abgebrochen. Starte ihn neu und versuche es erneut.";
        return prefix+": "+m;
    }
    private static String safe(Exception e){return e.getMessage()==null||e.getMessage().trim().isEmpty()?e.getClass().getSimpleName():e.getMessage();}
    private void closePending(String key){Pending p=pending.remove(key);if(p!=null)close(p.socket);}
    private void closeSession(String key){Session s=sessions.remove(key);if(s!=null)s.close();}
    private static void close(Socket s){if(s!=null)try{s.close();}catch(IOException ignored){}}

    private static final class Pending{
        final SSLSocket socket;final X509Certificate client,server;final long expires;
        Pending(SSLSocket s,X509Certificate c,X509Certificate t,long e){socket=s;client=c;server=t;expires=e;}
    }
    private static final class ServerTrust implements X509TrustManager{
        final String expected;final boolean allow;final AtomicReference<X509Certificate> captured;
        ServerTrust(String e,boolean a,AtomicReference<X509Certificate> c){expected=e;allow=a;captured=c;}
        @Override public void checkClientTrusted(X509Certificate[]chain,String auth){}
        @Override public void checkServerTrusted(X509Certificate[]chain,String auth)throws java.security.cert.CertificateException{
            if(chain==null||chain.length==0||!(chain[0].getPublicKey() instanceof RSAPublicKey))throw new java.security.cert.CertificateException("TV-Zertifikat ist ungültig.");
            if(captured!=null)captured.set(chain[0]);if(allow)return;
            try{String actual=fingerprint(chain[0]);if(expected==null||!MessageDigest.isEqual(expected.getBytes(StandardCharsets.US_ASCII),actual.getBytes(StandardCharsets.US_ASCII)))throw new java.security.cert.CertificateException("TV-Zertifikat stimmt nicht mit dem gespeicherten Zertifikat überein.");}
            catch(java.security.cert.CertificateException e){throw e;}catch(Exception e){throw new java.security.cert.CertificateException("TV-Zertifikat konnte nicht geprüft werden.",e);}
        }
        @Override public X509Certificate[]getAcceptedIssuers(){return new X509Certificate[0];}
    }    private static final class Session{
        final String key,host;final SSLSocket socket;final Set<String>caps=Collections.synchronizedSet(new HashSet<String>());final Object writeLock=new Object();
        volatile boolean closed;volatile int features,imeCounter,imeFieldCounter;volatile String model="",manufacturer="Android TV",version="";
        Session(String k,String h,SSLSocket s){key=k;host=h;socket=s;}
        boolean isOpen(){return !closed&&!socket.isClosed()&&socket.isConnected();}
        void send(byte[]raw)throws IOException{synchronized(writeLock){if(!isOpen())throw new IOException("TV-Verbindung ist geschlossen.");ProtoWire.writeDelimited(socket.getOutputStream(),raw);}}
        void buildCaps(){caps.clear();if((features&KEY)!=0){Collections.addAll(caps,"up","down","left","right","ok","back","home","menu","settings","play","pause","playPause","rewind","fastForward","channelUp","channelDown","info","guide","source");for(int i=0;i<=9;i++)caps.add("digit"+i);}
            if((features&POWER)!=0)caps.add("power");if((features&VOLUME)!=0)Collections.addAll(caps,"volUp","volDown","mute");if((features&IME)!=0)caps.add("text");}
        void startReader(){Thread t=new Thread(()->{try{while(isOpen()){List<ProtoWire.Field>fields=ProtoWire.parse(ProtoWire.readDelimited(socket.getInputStream()));for(ProtoWire.Field f:fields){
                    if(f.number==8&&f.data!=null){ProtoWire.Field v=ProtoWire.first(ProtoWire.parse(f.data),1);send(ProtoWire.remote(9,ProtoWire.varintField(1,v==null?0:v.integer)));}
                    else if(f.number==20&&f.data!=null){List<ProtoWire.Field>m=ProtoWire.parse(f.data);ProtoWire.Field app=ProtoWire.first(m,1),status=ProtoWire.first(m,2);if(app!=null){ProtoWire.Field n=ProtoWire.first(ProtoWire.parse(app.data),1);if(n!=null)imeCounter=(int)n.integer;}if(status!=null){ProtoWire.Field n=ProtoWire.first(ProtoWire.parse(status.data),1);if(n!=null)imeFieldCounter=(int)n.integer;}}
                    else if(f.number==22&&f.data!=null){ProtoWire.Field status=ProtoWire.first(ProtoWire.parse(f.data),2);if(status!=null){ProtoWire.Field n=ProtoWire.first(ProtoWire.parse(status.data),1);if(n!=null)imeFieldCounter=(int)n.integer;}}
                }}}catch(Exception ignored){close();}},"TrashyRemote-AndroidTV");t.setDaemon(true);t.start();}
        void close(){closed=true;AndroidTvRemoteAdapter.close(socket);}
    }
}