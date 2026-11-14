package de.trashygecko.tspiele;

import android.content.BroadcastReceiver;
import android.content.Context;
import android.content.Intent;

/**
 * Nach einem Neustart oder einem App-Update sind alle Wecker weg – hier werden
 * sie wieder eingeplant.
 */
public class LobStart extends BroadcastReceiver {

    @Override
    public void onReceive(Context kontext, Intent absicht) {
        LobWecker.plane(kontext);
    }
}
