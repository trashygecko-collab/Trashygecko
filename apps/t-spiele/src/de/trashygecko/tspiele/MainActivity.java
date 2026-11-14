package de.trashygecko.tspiele;

import android.app.Activity;
import android.content.Intent;
import android.net.Uri;
import android.os.Bundle;
import android.widget.Toast;

/**
 * T.Spiele hat keine eigene Oberfläche: beim Antippen des Symbols wird die
 * Connector-Adresse an den Browser übergeben, und die App schließt sich sofort
 * wieder. Mehr soll sie nicht tun.
 */
public class MainActivity extends Activity {

    /** Die Adresse des Connectors im Internet. */
    private static final String ADRESSE = "https://trashygecko.onrender.com/connector";

    @Override
    protected void onCreate(Bundle zustand) {
        super.onCreate(zustand);
        try {
            Intent absicht = new Intent(Intent.ACTION_VIEW, Uri.parse(ADRESSE));
            absicht.addFlags(Intent.FLAG_ACTIVITY_NEW_TASK);
            startActivity(absicht);
        } catch (Exception fehler) {
            // Kommt praktisch nur vor, wenn auf dem Gerät kein Browser ist.
            Toast.makeText(this, "Kein Browser gefunden – bitte einen installieren.",
                    Toast.LENGTH_LONG).show();
        }
        finish();
    }
}
