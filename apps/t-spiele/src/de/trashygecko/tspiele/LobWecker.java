package de.trashygecko.tspiele;

import android.app.AlarmManager;
import android.app.Notification;
import android.app.NotificationChannel;
import android.app.NotificationManager;
import android.app.PendingIntent;
import android.content.BroadcastReceiver;
import android.content.Context;
import android.content.Intent;
import android.os.Build;

import java.util.Calendar;
import java.util.Random;

/**
 * Schickt jeden Morgen um 7:30 ein Lob als Benachrichtigung – ohne Server,
 * ohne Internet: der Wecker steckt im Gerät selbst.
 */
public class LobWecker extends BroadcastReceiver {

    /** Kennung des Benachrichtigungskanals (ab Android 8 nötig). */
    public static final String KANAL = "morgenlob";

    private static final int KENNUNG = 7241;
    private static final int STUNDE = 7;
    private static final int MINUTE = 30;

    /** Lob-Sprüche ohne jeden Bezug auf ein bestimmtes Spiel. */
    private static final String[] LOBS = {
        "Guten Morgen! Du hast es drauf.",
        "Ein neuer Tag, eine neue Chance – du packst das!",
        "Heute wird ein guter Tag. Zeig es ihnen!",
        "Du bist heute nicht zu stoppen!",
        "Guten Morgen, Champion!",
        "Einfach spitze, dass du da bist!",
        "Der Tag wartet – und du bist bereit!",
        "Frisch ausgeschlafen und unschlagbar!",
        "Ein Lob zum Wachwerden: Du machst das super!",
        "Morgenlicht und gute Laune – leg los!",
        "Neuer Tag, neue Bestleistung!",
        "Augen auf, Kopf an, Spaß an: Guten Morgen!",
        "Der frühe Vogel spielt eine Runde. Los geht's!",
        "Guten Morgen – die Welt braucht dich wach!",
        "Weiter so, heute ist dein Tag!",
        "Zeit für eine Runde T.Spiele – guten Morgen!",
        "Zweimal gelobt: Guten Morgen und gut gemacht!",
        "Früh wach? Dann kann der Tag nur gut werden.",
        "Du machst das großartig – jeden Tag.",
        "Aufstehen lohnt sich: heute bist du dran!"
    };

    @Override
    public void onReceive(Context kontext, Intent absicht) {
        zeige(kontext);
        // Gleich den nächsten Morgen einplanen, damit die Reihe nicht abreißt.
        plane(kontext);
    }

    /** Legt den Wecker auf den nächsten Morgen um 7:30. */
    public static void plane(Context kontext) {
        AlarmManager wecker = (AlarmManager) kontext.getSystemService(Context.ALARM_SERVICE);
        if (wecker == null) return;

        int flags = PendingIntent.FLAG_UPDATE_CURRENT;
        if (Build.VERSION.SDK_INT >= 23) flags |= PendingIntent.FLAG_IMMUTABLE;

        PendingIntent termin = PendingIntent.getBroadcast(kontext, KENNUNG,
                new Intent(kontext, LobWecker.class), flags);

        Calendar zeit = Calendar.getInstance();
        zeit.set(Calendar.HOUR_OF_DAY, STUNDE);
        zeit.set(Calendar.MINUTE, MINUTE);
        zeit.set(Calendar.SECOND, 0);
        zeit.set(Calendar.MILLISECOND, 0);
        if (zeit.getTimeInMillis() <= System.currentTimeMillis()) {
            zeit.add(Calendar.DAY_OF_YEAR, 1);
        }

        if (Build.VERSION.SDK_INT >= 23) {
            // Weckt auch aus dem Ruhezustand (Doze).
            wecker.setAndAllowWhileIdle(AlarmManager.RTC_WAKEUP, zeit.getTimeInMillis(), termin);
        } else {
            wecker.set(AlarmManager.RTC_WAKEUP, zeit.getTimeInMillis(), termin);
        }
    }

    /** Kanal anlegen (ab Android 8) – mehrfaches Aufrufen schadet nicht. */
    public static void kanalAnlegen(Context kontext) {
        if (Build.VERSION.SDK_INT < 26) return;
        NotificationManager verwaltung =
                (NotificationManager) kontext.getSystemService(Context.NOTIFICATION_SERVICE);
        if (verwaltung == null) return;
        NotificationChannel kanal = new NotificationChannel(KANAL, "Morgen-Lob",
                NotificationManager.IMPORTANCE_DEFAULT);
        kanal.setDescription("Jeden Morgen um 7:30 ein Lob");
        verwaltung.createNotificationChannel(kanal);
    }

    private void zeige(Context kontext) {
        NotificationManager verwaltung =
                (NotificationManager) kontext.getSystemService(Context.NOTIFICATION_SERVICE);
        if (verwaltung == null) return;

        kanalAnlegen(kontext);

        int flags = PendingIntent.FLAG_UPDATE_CURRENT;
        if (Build.VERSION.SDK_INT >= 23) flags |= PendingIntent.FLAG_IMMUTABLE;
        PendingIntent tippen = PendingIntent.getActivity(kontext, 0,
                new Intent(kontext, MainActivity.class), flags);

        Notification.Builder bauen = Build.VERSION.SDK_INT >= 26
                ? new Notification.Builder(kontext, KANAL)
                : new Notification.Builder(kontext);

        bauen.setSmallIcon(android.R.drawable.ic_dialog_info)
             .setContentTitle("Guten Morgen!")
             .setContentText(zufall())
             .setContentIntent(tippen)
             .setAutoCancel(true);

        if (Build.VERSION.SDK_INT >= 21) {
            bauen.setPriority(Notification.PRIORITY_DEFAULT);
        }

        verwaltung.notify(KENNUNG, bauen.build());
    }

    private static String zufall() {
        return LOBS[new Random().nextInt(LOBS.length)];
    }
}
