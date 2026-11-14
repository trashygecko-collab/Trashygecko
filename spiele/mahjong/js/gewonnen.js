const GEWONNEN_SPRUECHE = [
    "a, wer ist denn hier das ganz große Genie? Hätte zwar fast die Realität gesprengt, aber hat geklappt! Ja, feini!",
    "Ab heute wird dein Name in Großbuchstaben geschrieben.",
    "Ab sofort darfst du deine eigenen Gesetze schreiben. Die alten scheinen dich ohnehin nicht aufzuhalten.",
    "Ach, du kleines Wunder auf zwei Beinen. Wer hätte dir das zugetraut?",
    "„Alle Paare gefunden! Cupid wäre stolz auf dich!“",
    "Alle Paare vereint. Was für ein Happy End!",
    "Alle sagten, das geht schief. Du hast es gemacht – und sie sprachlos zurückgelassen.",
    "Alle Steine weg! Jetzt kannst du endlich wieder klar sehen!",
    "Alles ist gut. Du hast so viel Energie reingesteckt, jetzt darfst du dich einfach mal zurücklehnen und stolz auf dich sein.",
    "Als Belohnung für diesen historischen Triumph darfst du heute den Müll rausbringen. Mit stolz geschwellter Brust!",
    "Andere brauchen einen Plan. Du brauchst offenbar nur ausreichend Selbstvertrauen.",
    "Andere gehen spazieren. Ich gehe Mahjong spielen.",
    "Angesichts der absoluten Unvermeidbarkeit des Scheiterns hast du mit unerbittlicher, kalter Entschlossenheit die Realität selbst gebogen. Wir verneigen uns in eisiger Ehrfurcht.",
    "Atme erst mal tief aus, mein Schatz. Das Schlimmste ist vorbei, und du hast das so unfassbar gut gemacht.",
    "Atmen wir erst einmal tief aus. Du hast in den letzten Stunden so unglaublich viel Energie, Nerven und Kraft investiert, und jetzt darfst du den ganzen Druck einfach komplett von dir abfallen lassen. Du bist angekommen, und das Schlimmste ist geschafft.",
    "Auch kleine Schritte können riesige Siege sein. Und deiner zählt heute ganz besonders.",
    "Auch wenn der Weg dorthin steinig war und es sich manchmal wie ein Drahtseilakt angefühlt hat: Du bist durchgegangen, und das zeigt, wie unerschütterlich stark du bist.",
    "Aufgeben ist keine Option. Höchstens eine kurze Kaffeepause mit dramatischem Blick aus dem Fenster.",
    "Augenzeugen berichten von ungläubigem Staunen und spontanen Standing Ovations.",
    "Ausgezeichnet mit dem silbernen Rettungsring für außergewöhnliche Selbstrettung.",
    "Bei dir weiß man nie, ob man helfen, lachen oder vorsichtshalber die Feuerwehr rufen soll.",
    "Bescheidenheit ist heute leider ausverkauft.",
    "Bitte einmal kurz aufstehen. Eine Legende hat gerade den Raum betreten.",
    "Chapeau, ich dachte wirklich, das wird der Untergang, aber du hast es gerockt.",
    "Da hat der kleine Strolch doch tatsächlich das Chaos bezwungen, bevor die Erwachsenen panisch die Evakuierung starten mussten. Fein!",
    "Dafür darfst du dich jetzt exakt fünf Minuten lang wie ein Gott fühlen, bevor dir auffällt, dass der Müll immer noch in der Küche steht.",
    "Dafür darfst du dir heute Abend im Supermarkt ein Joghurt deiner Wahl aussuchen – aber nur mit rotem Rabattsticker.",
    "Dafür darfst du dir jetzt eine halbe Stunde lang einreden, du hättest das von Anfang an genau so geplant.",
    "Dafür darfst du heute Abend ausnahmsweise bestimmen, welches YouTube-Video im Hintergrund läuft, während du erschöpft in die Kissen sinkst.",
    "Dafür kriegst du einen Orden. Aus Pappe. Der bei Luftfeuchtigkeit sofort zerfällt.",
    "Dafür musst du heute Nacht wenigstens nicht im Treppenhaus schlafen. Auf der Fußmatte ist noch Platz.",
    "Dafür springt heute Abend maximal ein lauwarmes Bier und ein Schulterklopfen raus, das leicht nach Mitleid schmeckt.",
    "Das Ergebnis ist gut. Über den Weg dorthin reden wir ein anderes Mal.",
    "Das hat mehr Nerven gekostet als eine Geiselnahme, aber das Ergebnis entschädigt für alles.",
    "Das hat so viel Nerven gekostet, dass die nächste Therapiesitzung auf dein Ticket geht.",
    "Das war der endgültige Beweis, dass man mit einer gesunden Portion Wahnsinn im Leben alles erreichen kann.",
    "Das war der perfekte Beweis dafür, dass man mit genügend Chaos immer ans Ziel kommt.",
    "Das war der ultimative Beweis: Wer keine Ahnung hat, improvisiert sich einfach souverän an die Spitze.",
    "Das war dermaßen unerwartet, dass ich kurz meine Berechnungen überprüfen musste.",
    "Das war ein Glanzmoment für die Menschheit. Oder zumindest für den Blutdruck deines Hausarztes.",
    "Das war ein Husarenstreich, der so haarscharf an einer Kernkatastrophe vorbeigeschrammt ist, dass es fasziniert.",
    "Das war ein Husarenstück, das selbst Baron von Münchhausen vor Neid erblassen lassen würde.",
    "Das war ein Ritt auf der Rasierklinge, aber du hast nicht mal einen Kratzer davongetragen.",
    "„Das war eine runde Sache. Obwohl die Steine eckig waren!“",
    "Das war eine steinige Geburt – aber ein Happy End!",
    "Das war fachlich vielleicht fragwürdig, menschlich aber eine Glanzleistung.",
    "Das war haarscharf am Nervenzusammenbruch vorbei, aber das Ergebnis kann sich sehen lassen.",
    "Das war kein einfacher Schritt nach vorn, das war ein gewaltiger Sprung über den eigenen Schatten.",
    "Das war kein gewöhnlicher Erfolg. Das war ein dreistündiger Blockbuster mit Happy End.",
    "Das war knapp, aber knapp ist bekanntlich auch vorbei – und vorbei ist gewonnen!",
    "Das war nicht nur gut. Das war stein-gut!",
    "Das war nicht unbedingt intelligent, aber verdammt effektiv.",
    "Das war so brillant und dreist, dass es fast schon wieder verdächtig ist.",
    "Das war so gut, dass es fast unfair gegenüber allen anderen ist, die es gar nicht erst versucht haben.",
    "Das war so gut, dass sogar die Steine applaudieren würden. Wenn sie Hände hätten.",
    "Das war so knapp am Totalverlust vorbei, dass dein Schutzengel jetzt offiziell Burnout angemeldet hat und in Frührente geht.",
    "Das war so nah an der Katastrophe, dass es schon wieder moderne Kunst war.",
    "Das war so professionell, das könnte im Lebenslauf glatt als „Chaos-Management mit Todesfolge“ durchgehen.",
    "Das war so souverän, dass selbst die hartnäckigsten Skeptiker ihre Popcorn-Tüten weggelegt haben.",
    "Das war steinstark!",
    "Dein Denkmal wird selbstverständlich aus purem Gold gefertigt.",
    "Dein Erfolg ist in Stein gemeißelt!",
    "Dein Erfolg widerspricht mindestens drei wissenschaftlichen Theorien und einer Bedienungsanleitung.",
    "Dein Gehirn arbeitet offenbar mit einer Software, für die es noch keine Updates gibt.",
    "Dein innerer Kompass zeigt zwar gelegentlich Richtung Frittenbude, aber heute hat er den richtigen Weg gefunden.",
    "Dein königlicher Titel lautet: Herrscher über Chaos, Zufall und gelegentliche Geistesblitze.",
    "Denk immer daran: Irgendwo gibt es jemanden, der noch weniger Ahnung hat als du. Und der hat heute frei.",
    "Der Ausschuss empfiehlt dringend eine angemessene Selbstbewunderung.",
    "Die Börse reagiert positiv, nachdem dein Erfolg offiziell bestätigt wurde.",
    "Die Erwartungen waren im Keller, aber du hast sie von unten noch untergraben – im positiven Sinne!",
    "Die Geschichtsbücher werden diesen Moment ignorieren, aber wir wissen, was du da vollbracht hast.",
    "Die Jury war sich uneinig, ob sie applaudieren oder einen Krankenwagen bestellen sollte.",
    "Die Kommission zur Feststellung historischer Großtaten hat getagt: Deine Bewältigung dieses kolossalen Desasters wird als unumstößlicher Meilenstein in die Annalen des Überlebens eingehen.",
    "Die Krone sitzt. Der Verstand ist optional. Der Erfolg ist amtlich.",
    "Die Naturgesetze schrien in Agonie, die Statistik forderte deinen sofortigen Rücktritt – und dennoch steht das Werk. Dies ist kein bloßer Erfolg, das ist ein Triumph über die Logik.",
    "„Die Paarvermittlung war erfolgreich. Kein Stein muss heute allein schlafen!“",
    "Die Statistik hat dagegen gesprochen, aber dein Sturkopf hat gewonnen.",
    "Die Steine haben kapituliert. Du darfst dich jetzt wichtig fühlen.",
    "Die Steine hatten keine Chance. Du aber offenbar jede Menge Geduld!",
    "Die Steine sind weg. Dein Grinsen darf bleiben!",
    "Die Steine sind weg. Jetzt musst du nur noch deine Kaffeetasse finden.",
    "Die Überlebenschancen waren gering, aber dein Ego hat es am Ende gerichtet.",
    "Die Wahrscheinlichkeit deines Erfolgs wurde gerade wegen mangelnder Beweise aus der Statistik entfernt.",
    "Die Wahrscheinlichkeit sprach dagegen, aber wen kümmert schon Mathe, wenn man Talent hat?",
    "Die Welt hat einen neuen Maßstab für außergewöhnliche Leistungen.",
    "Die Würde des Augenblicks verbietet jeden billigen Jubel. Stattdessen gebührt dir das schweigende, unumstößliche Anerkennen einer monumentalen, fast schon beängstigenden Kompetenz.",
    "Die Zivilisation stand am Rande des Abgrunds, bereit, sich selbst aufzugeben. Du hast nicht nur eingegriffen, du hast das Schicksal eigenhändig zur Raison gebracht.",
    "Die zuständige Behörde ist sprachlos und hat den Vorgang geschlossen.",
    "Du bist der Grund, warum Bedienungsanleitungen inzwischen Warnhinweise haben.",
    "Du bist der lebende Beweis dafür, dass man mit genug Frechheit und Können jedes System sprengen kann.",
    "Du bist der lebende Beweis, dass man auch mit einer Bedienungsanleitung auf dem Kopf erfolgreich sein kann.",
    "Du bist ein echtes Geschenk, und ich freue mich einfach riesig mit dir über diesen Erfolg.",
    "Du bist eine Zumutung für die Logik und eine Bereicherung für die Erfolgsgeschichte.",
    "Du bist heute der ungekrönte König der Zwillingssuche!",
    "Du bist manchmal ein wandelnder Systemfehler, aber heute einer mit hervorragendem Ergebnis.",
    "Du bist manchmal eine Herausforderung für die Menschheit, aber heute eine verdammt erfolgreiche.",
    "Du bist nicht gescheitert. Du hast lediglich eine alternative Erfolgsroute entdeckt.",
    "Du bist nicht nur dann wunderbar, wenn dir alles gelingt. Du bist es auch an den Tagen, an denen du einfach nur weitermachst.",
    "Du bist wie ein Phoenix aus der Aschenbahn der totalen Inkompetenz aufgestiegen.",
    "Du bist wunderbar, genau so wie du bist. Und heute hast du der Welt (und dir selbst) wieder gezeigt, was für ein toller Mensch in dir steckt.",
    "Du darfst jetzt weich werden und nachgeben. Die Anspannung war riesig, aber du hast es getragen. Ganz allein.",
    "Du darfst jetzt weich werden, nachgeben und stolz auf dich sein. Mach dir keinen Kopf mehr über das, was hätte schiefgehen können – es ist gut ausgegangen, weil du nicht aufgegeben hast.",
    "Du hast alle Paare gefunden. Jetzt fehlt nur noch der passende Lottoschein!",
    "Du hast allen das Gegenteil bewiesen, die schon popcornbereit im Stuhl saßen.",
    "Du hast das Ding gerockt, während der Rest der Welt noch panisch die Anleitung gesucht hat.",
    "Du hast das Ding so genial gedreht, dass die Physik kurz überlegt hat, ihre eigenen Gesetze zu ändern.",
    "Du hast das Ding so souverän gewuppt, als hättest du nie etwas anderes getan (außer Panik zu schieben).",
    "Du hast das Ding überlebt und dabei sogar noch geschafft, gut auszusehen.",
    "Du hast das Problem nicht verstanden, aber ihm trotzdem erfolgreich den Garaus gemacht.",
    "Du hast das Ruder im allerletzten Moment herumgerissen, als das Schiff schon lichterloh brannte.",
    "Du hast das Schiff durch den Sturm manövriert, obwohl der Kapitän (du) absolut keinen Plan hatte.",
    "Du hast das Schlachtfeld betreten, alles niedergerissen und stehst jetzt auf den Trümmern – erfolgreich!",
    "Du hast das sprichwörtliche Brett vorm Kopf abgenommen und direkt damit das Tor eingeschlagen.",
    "Du hast das System komplett ausgetrickst und bist jetzt der ungekrönte Star des Tages.",
    "Du hast das Unmögliche geschafft, und jetzt trinken wir kollektiv auf deine absolut schamlose Sturheit.",
    "Du hast das Unmögliche geschafft, und jetzt tun wir alle so, als wäre das eine Kleinigkeit gewesen.",
    "Du hast das Unmögliche geschafft, und jetzt verneigen sich selbst die Kritiker, die schon mit dem Grabstein wedelten.",
    "Du hast das Unmögliche möglich gemacht und siehst dabei noch so aus, als ob es ein öder Mittwoch wäre.",
    "Du hast das Unmögliche möglich gemacht, und jetzt müssen wir alle so tun, als wäre das normal.",
    "Du hast das Ziel erreicht, obwohl der Pfad seit Jahren von Dornen und Verzweiflung überwuchert war.",
    "Du hast dem drohenden Untergang direkt ins Gesicht gelacht und gewonnen.",
    "Du hast dem finsteren Abgrund direkt angelächelt und bist einfach drübergesprintet.",
    "Du hast den Berg erklommen, den alle anderen schon längst weggesprengt hatten.",
    "Du hast den Berg erklommen, obwohl jeder wusste, dass es dort oben keinen Sauerstoff gibt.",
    "Du hast den Ernst der Lage erkannt und sie prompt ignoriert – mit vollem, glorreichen Erfolg.",
    "Du hast den gordischen Knoten einfach durchgehackt, nachdem alle anderen daran verzweifelt sind.",
    "Du hast den gordischen Knoten nicht nur gelöst, sondern ihn direkt feierlich verbrannt.",
    "Du hast den Hauptgewinn geknackt, obwohl du anfangs nicht mal wusstest, wo der Automat steht.",
    "Du hast den Karren nicht nur aus dem Dreck gezogen, sondern gleich noch ein Fluchtfahrzeug draus gemacht.",
    "Du hast den Karren so tief in den Sand gesetzt, dass du am anderen Ende der Welt wieder herausgekommen bist.",
    "Du hast den Kopf im letzten Moment aus der Schlinge gezogen und den Strick direkt als stylische Krawatte umgebunden.",
    "Du hast den Nagel so präzise auf den Kopf getroffen, dass der Hammer in tausend Teile zersprungen ist.",
    "Du hast den Sack zugemacht, bevor überhaupt jemand gemerkt hat, dass die Katastrophe entwischt war.",
    "Du hast den ultimativen Bosskampf im Real Life einfach ohne Cheat-Codes durchgespielt.",
    "Du hast die Messlatte so unfassbar hoch gelegt, dass wir sie wohl nie wieder finden werden.",
    "Du hast die Mission erfüllt, obwohl die Zielscheibe die ganze Zeit auf deinem eigenen Rücken klebte.",
    "Du hast die Steine ganz schön ins Rollen gebracht!",
    "Du hast dir so viele Sorgen gemacht und nachts gegrübelt – und sieh nur, wie toll du das am Ende gelöst hast.",
    "Du hast es durchgezogen wie ein eiskalter Profi, der heimlich im Keller geübt hat.",
    "Du hast es geschafft – und jetzt atmen wir alle tief aus und trinken einen auf deine Nerven aus Stahl.",
    "„Du hast es geschafft! Die Steine sind sprachlos. Und das will bei Steinen etwas heißen!“",
    "Du hast es geschafft, ohne dass wir uns seelisch auch nur ansatzweise darauf einstellen konnten.",
    "Du hast es geschafft, und das Beste daran ist: Niemand kann jetzt behaupten, du hättest auch nur einen Hauch von Hilfe gehabt.",
    "Du hast es geschafft, und das Schlimmste daran ist: Es sieht auch noch verdammt gut aus.",
    "Du hast geliefert, als hätte dein Leben davon abgehängt – oder zumindest dein sehr stolzer Ruf.",
    "Du hast heute eindeutig den richtigen Stein im Brett!",
    "Du hast heute nicht nur ein Problem gelöst, sondern auch ein Stück Vertrauen in dich selbst zurückgewonnen.",
    "Du hast schon so viel geschafft – lass dich jetzt nicht unterkriegen. Atme durch, einmal kurz schütteln, und dann weiter Schritt für Schritt. Du bist stärker, als du denkst. Du schaffst das!",
    "Du hast sie alle erwischt. Die Steine werden sich rächen!",
    "Du hast so oft an dir gezweifelt, aber deine Ausdauer hat gewonnen. Du bist einfach bewundernswert.",
    "Du hast zwar keine Ahnung gehabt, aber dafür eine beeindruckende Konsequenz an den Tag gelegt.",
    "Du kannst so verdammt stolz auf dich sein. Ich bin es auf jeden Fall, und zwar von ganzem Herzen.",
    "Du musst nicht immer stark sein. Aber heute darfst du ruhig stolz auf deine Stärke sein.",
    "Du schaffst das! Und wenn nicht, behaupten wir einfach, es war ein Test.",
    "Egal, was alle anderen sagen: Ich habe gesehen, wie hart du dafür gekämpft hast. Und das Ergebnis ist einfach wunderbar.",
    "Ehrenmitglied der Gesellschaft zur Verhinderung unnötiger Katastrophen.",
    "Ehrlich gesagt dachte ich, das endet im Desaster, aber chapeau, du hast abgeliefert.",
    "Ei, ei, ei – fast hätt's gekracht, aber du hast den Kopf aus der Schlinge gezogen. Wer kriegt jetzt ein Gummibärchen? Du!",
    "Ei, ei, ei, das war haarscharf vor dem Nervenzusammenbruch, aber jetzt wird geklatscht! Fein gemacht!",
    "Ei, ei, ei, wer hat das Unmögliche möglich gemacht und sieht dabei noch völlig unschuldig aus? Du!",
    "Ei, ei, ei, wer hat den fetten Patzer einfach durch freches Weitermachen ausgebügelt? Du kleiner Rebell!",
    "Ei, guck mal an! Ganz alleine den Karren in den Dreck gefahren und im letzten Moment wieder rausgezogen. Wer ist ein kluges Köpfchen? Du!",
    "Ei, schau an, der kleine Meister des Notfalls hat wieder zugeschlagen. Wer ist ein Genie? Du!",
    "Ei, schau an, der kleine Überlebenskünstler hat den Bosskampf im Real Life ohne Cheat geschafft. Fein!",
    "Ei, schau mal einer an, der kleine Tollpatsch ist im Ziel angekommen und stolpert direkt über den Siegerpokal!",
    "Ei, schau mal einer an: Vom Totalausfall zum Überlebenskünstler in drei Sekunden. Wer ist ein braves Kind?",
    "Ei, schau mal, der kleine Chaos-König hat den Laden gerettet. Wer hätte das gedacht?",
    "Ei, schau mal, die Welt existiert noch, weil du im letzten Moment den Daumen draufgehalten hast. Feines Kind!",
    "Ei, wer hat den Fehler behoben, ohne dass das System kollabierte? Ja, du!",
    "Ei, wer hat den Karren so tief reingefahren und ihn dann doch noch rausgezogen? Eine Meisterleistung der Verwirrung!",
    "Ei, wer hat denn das schwere Ding ganz alleine geschafft, ohne dass echte Tränen geflossen sind? (Fast-Tränen zählen nicht!)",
    "Ei, wer hat denn dem drohenden Untergang so tapfer in die Augen geblickt? Du kleiner Draufgänger!",
    "Ei, wer ist denn hier der ungeschlagene Champion im Haarscharf-am-Desaster-Vorbeischrammen? Du!",
    "Ei, wer kriegt jetzt ein Fleißbändchen für das Verhindern des totalen Ruins? Du natürlich!",
    "Eigentlich war das absolut zum Scheitern verurteilt, aber du hast den Fluch gebrochen.",
    "Eilmeldung: Ein Mensch hat tatsächlich etwas zu Ende gebracht. Die Bevölkerung jubelt.",
    "Ein Anfänger fragt einen Mahjong-Meister: „Was ist das Geheimnis eines guten Spielers?“ Der Meister antwortet: „Ruhe, Geduld und ein gutes Gedächtnis.“ Der Anfänger nickt und fragt: „Und was ist, wenn man nichts davon hat?“ Der Meister lächelt: „Dann braucht man einen sehr langen Nachmittag.“",
    "Ein bisschen bekloppt, aber dafür außergewöhnlich erfolgreich.",
    "Ein denkwürdiger Moment: Du hast etwas Ernstes angefangen und es tatsächlich zu einem Ende gebracht.",
    "Ein dreifaches Hoch auf den Bezwinger der völlig unnötigen Katastrophen!",
    "Ein Geniestreich, der extrem knapp an der legalen Grenze vorbeigeschrammt ist.",
    "Ein grandioser Sieg, der exakt so geplant war – zumindest in deinen kühnsten Albträumen.",
    "Ein historischer Sieg! Die Belohnung dafür: Die Gewissheit, dass du morgen genau denselben Blödsinn wieder machst.",
    "Ein historischer Triumph der reinen Willenskraft über den gesunden Verstand.",
    "Ein klassischer Fall von: Niemand wusste, wie es geht – außer dir, auch wenn du es erst im Nachhinein gemerkt hast.",
    "„Ein Level weniger, ein Mahjong-Meister mehr!“",
    "Ein Mahjong-Spieler geht zum Psychologen. „Herr Doktor, ich habe ein Problem. Ich sehe überall Paare.“ Der Psychologe fragt: „Seit wann haben Sie das?“ „Seit ich angefangen habe, Mahjong zu spielen.“",
    "Ein Mahjong-Spieler wird gefragt: „Wie schaffst du es eigentlich, immer den Überblick über so viele Steine zu behalten?“ Er antwortet: „Ganz einfach. Ich merke mir, wo jeder Stein liegt.“ „Und wenn du es vergessen hast?“ „Dann suche ich eben so lange, bis ich ihn wiedergefunden habe!“",
    "Ein Mahjong-Stein fragt den anderen: „Hast du deinen Zwilling gesehen?“ Der andere antwortet: „Nein, aber ich hoffe, er hat wenigstens ein gutes Gedächtnis!“",
    "Ein Mahjong-Stein wird gefragt: „Warum liegst du denn ganz allein hier?“ „Mein Zwilling wurde gefunden.“ „Das ist doch schön!“ „Ja, aber leider mit einem anderen Stein.“",
    "Ein Mann spielt Mahjong und schaut sich lange die Steine an. Seine Frau fragt: „Was machst du da?“ „Ich trainiere mein Gedächtnis.“ „Und funktioniert es?“ „Was funktioniert?“",
    "Ein Mann spielt Mahjong und wird gefragt: „Ist das Spiel eigentlich schwierig?“ Er antwortet: „Nein, überhaupt nicht. Man muss nur die gleichen Steine finden.“ „Und warum sitzt du dann schon seit zwei Stunden davor?“ „Weil die Steine offenbar nicht wissen, dass sie gleich sein sollen!“",
    "Ein Meisterwerk der Improvisation, das während der Ausführung mehrfach kurz vor dem Einsturz stand.",
    "Ein modernes Wunder ist geschehen und du bist der alleinige Grund dafür.",
    "Ein monumentaler Sieg gegen jegliche Gesetze des gesunden Menschenverstandes.",
    "Ein Stein liegt ganz allein auf dem Spielfeld. Kommt ein anderer vorbei und fragt: „Warum bist du so traurig?“ Antwortet er: „Meine bessere Hälfte ist schon wieder verschwunden!“",
    "Ein Triumph sondergleichen, der dich vermutlich mindestens drei Jahre Lebenszeit gekostet hat.",
    "Eine Leistung, die selbst Hollywood nur mit Spezialeffekten glaubwürdig darstellen könnte.",
    "Eine neue Studie bestätigt: Du bist zu 73 Prozent Genie und zu 27 Prozent unberechenbares Risiko.",
    "Einfach legendär, wie du alle Zweifel mit puren Fakten und Talent erschlagen hast.",
    "Erhebt euch! Seine Majestät der Problemlösung hat gesprochen.",
    "Es ist völlig normal, dass du jetzt erschöpft bist. Du hast Großes geleistet und darfst dich jetzt feiern lassen.",
    "Es sah stundenlang nach Kernschmelze aus, aber du hast das Ding noch gerettet.",
    "Es war ein schwerer Weg und du hast gezweifelt, aber du bist durchgegangen. Das zeigt einfach, wie stark du bist.",
    "Fein gemacht! Fast hätten wir den Notarzt gebraucht, aber du hast den richtigen Knopf gedrückt. Tschakka!",
    "Fein, fein! Fast wäre es ein Desaster geworden, aber du hast die Kurve noch gekratzt. Brav!",
    "Forscher haben dein Vorgehen untersucht und anschließend beschlossen, lieber etwas anderes zu studieren.",
    "Für diese Leistung werden Ihnen rückwirkend sämtliche Zweifel erlassen.",
    "Für diese Meisterleistung kriegst du morgen früh ein extra weiches Ei. Aus Plastik.",
    "Geschafft! Alle Steine sind vom Tisch!",
    "Gewinner des goldenen Schraubenschlüssels für außergewöhnliche Improvisation.",
    "Gewinner des Preises für die kreativste Umgehung des gesunden Menschenverstands.",
    "Glückwunsch! Du hast mehr Steine bewegt als ein Umzugsunternehmen!",
    "Große Leistungen beginnen oft dort, wo andere längst aufgegeben haben.",
    "Guck an, der kleine Pechvogel hat sich einfach zum Glückspilz hochgestolpert. Ei, wie niedlich!",
    "Guck mal, der kleine Wirbelwind hat den Weltuntergang einfach mal um eine Woche verschoben. Ei, wie süß!",
    "Hätte absolut niemand drauf gewettet, aber du hast es tatsächlich durchgezogen.",
    "Herzlichen Glückwunsch! Deine Augen funktionieren noch hervorragend!",
    "„Herzlichen Glückwunsch! Du hast den Steinen gezeigt, wo der Hammer hängt!“",
    "Hiermit wird Ihnen die uneingeschränkte Befähigung zur Problembewältigung bescheinigt.",
    "Hui! Da staunt der Laie und der Fachmann weint – aber du hast es geschafft, mein kleiner Chaot!",
    "Hui, da hat aber jemand gezittert! Aber das Ergebnis zählt. Wer ist ein tapferes Kind?",
    "Hui, da klopfen wir uns aber mal auf die Schulter! Der kleine Überlebenskünstler hat's mal wieder geschafft.",
    "Hui, das war ein Ritt auf der Rasierklinge, aber du hast nicht mal einen Kratzer. Brav!",
    "Hui, das war knapp an der Kernschmelze vorbei! Aber wer hat's gerettet? Unser kleiner Problemlöser!",
    "Hui, wer hat denn da den Mega-Fehler ausgebügelt, bevor es jemand bemerkt hat? Unser kleiner Ninja!",
    "Hut ab, du hast aus dem drohenden Totalausfall einen absoluten Blockbuster gemacht.",
    "Hut ab, du hast das Ding nach Hause gebracht, obwohl der Kompass längst geschmolzen war.",
    "Ich bin beeindruckt. Und das passiert mir wirklich nicht oft.",
    "Ich bin einfach so unglaublich stolz auf dich – nicht nur, weil es geklappt hat, sondern weil ich weiß, wie viel Kraft dich das gekostet hat.",
    "Ich bin mir sicher, die Evolution hat sich bei dir einfach gedacht: „Ach, lass ihn machen, Physik und Logik interessieren ihn eh nicht.“",
    "Ich drücke dich in Gedanken ganz fest. Du hast das so sauber und mit so viel Herz durchgezogen.",
    "Ich hätte dir ja geholfen, aber offenbar war meine Anwesenheit völlig überflüssig.",
    "Ich hoffe, du siehst dich für einen Moment mit meinen Augen. Dann würdest du verstehen, wie viel du wert bist.",
    "Ich könnte dich gerade gleichzeitig knuddeln und für deinen Lösungsweg ausschimpfen.",
    "Ich möchte deine Vorgehensweise nicht hinterfragen. Ich möchte sie einfach nur überleben.",
    "Ich möchte nicht übertreiben, aber die NASA hat gerade nach deinem Lösungsweg gefragt.",
    "Ich war kurz davor, mein Testament zu machen, als du angefangen hast, aber du hast es tatsächlich überlebt.",
    "Ich würde sagen, das war geplant. Wir müssen ja nicht alles verraten.",
    "Ich würde sagen, du hast einen an der Waffel. Aber offenbar ist genau das dein Erfolgsgeheimnis.",
    "Ich wusste doch, dass unter dem ganzen Chaos ein kleines Genie wohnt.",
    "Ich wusste zu jeder Sekunde, dass du das schaffst – auch dann, als du selbst nicht mehr daran geglaubt hast.",
    "Ihr Antrag auf Anerkennung Ihrer Genialität wurde einstimmig genehmigt.",
    "Ihr persönlicher Erfolgsstatus wurde auf „nicht mehr zu toppen“ hochgestuft.",
    "Ja, ei, ei, da guck her: Ein echtes kleines Wunder, weil die Wahrscheinlichkeit für diesen Erfolg bei exakt null lag. Toll gemacht!",
    "Ja, ei, wer hat denn da allen gezeigt, dass man mit purer Sturheit jede Statistik schlägt? Du kleiner Sturkopf!",
    "Ja, fein gemacht! Das hätte auch sehr, sehr böse enden können, aber Glück muss der Mensch haben.",
    "Ja, fein gemacht! Erst alles anzünden und dann geschickt löschen. Wer ist ein kluges Köpfchen?",
    "Ja, fein! Der Abgrund hat gerufen, aber du hast ihm einfach die Zunge rausgestreckt und gewonnen.",
    "Ja, fein! Wer hat den Karren aus dem Sumpf gezogen und sich dabei nur minimal dreckig gemacht? Du!",
    "Ja, feini gemacht! Der Untergang war nah, aber deine Ignoranz war stärker. Glückwunsch!",
    "Ja, feini gemacht! Wer hat den gordischen Knoten durchgehackt und dabei so getan, als wüsste man, was man tut? Du!",
    "Ja, feini! Fast in die Hose gegangen, aber am Ende glanzvoll gerettet. Wer ist der Beste? Du!",
    "Ja, huch! Da hat der kleine Strolch tatsächlich geliefert, als es drauf ankam. Feini, feini!",
    "Ja, huch! Da staunt man aber, dass das tatsächlich funktioniert hat. Wer ist ein schlaues Köpfchen?",
    "Ja, huch! Wer hat denn da das unlösbare Problem gelöst, ohne dass das Universum implodiert ist? Du feiner Spatz!",
    "Ja, huch, es hat geklappt! Ohne dass wir das Gebäude räumen mussten. Da machen wir drei Kreuzchen!",
    "Ja, huch, es ist vollbracht! Und niemand musste weinen (außer der Computer). Brav!",
    "Ja, wer hat denn da den gordischen Knoten zerfetzt und sich gefreut wie ein Schneekönig? Du!",
    "Ja, wer hat denn da so schön improvisiert, dass selbst die Physik kurz kapituliert hat? Du kleiner Magier!",
    "Ja, wer hat denn da so toll den Kopf aus der Schlinge gezogen? Dafür gibt's ein virtuelles Kügelchen Eis!",
    "Ja, wer ist denn der Champion des Hauses nach diesem heroischen Überlebenskampf? Du!",
    "Ja, wer kann denn da so toll Tasten drücken, dass die Katastrophe ausbleibt? Ja, du bist es!",
    "Jetzt hast du den Dreh raus. Oder besser gesagt: den Stein raus!",
    "Keine Ahnung, wie du das angestellt hast, ohne dass die Realität kollabiert ist, aber Respekt.",
    "Komm erst mal an, lehn dich zurück und lass den ganzen Druck von dir abfallen. Du hast es geschafft.",
    "Komm her, du verrückter Schatz. Das hast du wirklich toll gemacht.",
    "Kopf hoch, Brust raus, Bauch rein – und jetzt tu so, als hättest du einen Plan!",
    "Lehn dich jetzt zurück, genieß diesen Moment der Ruhe und lass dich von dem Gedanken tragen: Du hast das wunderbar gemacht. Und heute Abend hast du dir jede Form von Erholung, etwas Gutes zu essen und absolute Entspannung so was von verdient.",
    "Level geschafft! Kurz mal weg vom Bildschirm, bevor du anfängst, die Fliesen im Bad zu sortieren.",
    "Mach dir jetzt einen schönen Abend, genieß diesen Moment und lass dir gesagt sein: Du bist einfach spitze, und ich bin unendlich froh, dass es dich gibt.",
    "Macht Platz! Hier kommt jemand, der sogar dem Zufall Anweisungen gibt.",
    "Man fragt sich manchmal, ob du ein verkanntes Genie oder einfach nur lebensmüde bist, aber das Ergebnis stimmt.",
    "Man hat schon viele Helden fallen sehen, doch wer aus dem Auge des absoluten Sturms mit einem funktionierenden Ergebnis hervorgeht, verdient keinen Applaus, sondern ein Denkmal aus kaltem Marmor.",
    "Man hätte dich für diesen waghalsigen Alleingang einweisen müssen, aber es hat funktioniert.",
    "Man möge dir Weintrauben reichen und sämtliche Aufgaben von dir fernhalten.",
    "Man muss dir eins lassen: Wenn du etwas an die Wand fährst, baust du wenigstens direkt ein Kunstwerk daraus.",
    "Man nennt es Talent, du nennst es Glück, ich nenne es illegale Hexerei.",
    "Man soll den Tag nicht vor dem Abend loben, aber du hast ihn heute komplett rasiert.",
    "Man stolpert manchmal über den Erfolg, aber du bist mit Anlauf reingesprungen und hast darin gebadet.",
    "Manchmal ist Durchhalten die größte Form von Intelligenz.",
    "Manchmal muss man erst alles anzünden, um zu sehen, dass man es auch hätte retten können. Du hast es gerettet!",
    "Mission erfüllt! Alle Steine wieder in geordnete Verhältnisse gebracht!",
    "Nach eingehender Prüfung wurde festgestellt: Das war tatsächlich nicht schlecht.",
    "Nebenwirkung dieses Levels: Du versuchst gleich, im Supermarkt zwei identische Konservendosen wegzuklicken.",
    "Nicht der Weg macht den Sieger aus, sondern der Mut, ihn trotz aller Zweifel zu gehen.",
    "Nicht jeder Triumph ist laut. Manche bestehen einfach darin, nicht aufgegeben zu haben.",
    "Niemand hat auch nur einen Cent auf dich gesetzt, und jetzt räumst du das Feld ab.",
    "Niemand hat mit dir gerechnet – vor allem du selbst nicht –, aber du hast es geschafft.",
    "Niemand muss dir mehr beweisen, was in dir steckt, denn du hast es gerade allen – und vor allem dir selbst – eindrucksvoll gezeigt.",
    "Niemand muss dir mehr beweisen, was in dir steckt, du hast es dir selbst und allen anderen gerade eindrucksvoll gezeigt.",
    "Respekt! Nicht einmal die Steine konnten dich aufhalten!",
    "Respekt, du hast das Unmögliche geschafft und dabei nicht mal eine Miene verzogen.",
    "Selbst das Universum schien gegen dich zu arbeiten, aber du hast es trotzdem bezwungen.",
    "Selbst dein innerer Schweinehund hat gerade fristlos gekündigt.",
    "Selbst die Götter im Olymp staunen gerade neidisch über das, was du da abgeliefert hast.",
    "Selbst Murphys Gesetz hat vor deiner chaotischen Genialität kapituliert.",
    "Selbst nach dem gefühlten zehnten Herzinfarkt hast du es noch ins Ziel gebracht.",
    "Sensationelle Szenen spielen sich ab. Experten sprechen von einem historischen Durchbruch.",
    "Sieh mich an: Du hast es hinter dir und es ist gut ausgegangen. Jetzt darfst du dich einfach nur ausruhen.",
    "So viel Glück kann man eigentlich gar nicht haben. Du hast es trotzdem geschafft.",
    "So viel Köpfchen gehört eigentlich ausgezeichnet!",
    "So viel Ordnung auf einem Haufen Steine. Marie Kondo wäre begeistert!",
    "Sondersendung: Du hast geliefert. Weitere Informationen liegen noch nicht vor.",
    "Spürst du, wie die Anspannung abfällt? Genau so fühlt sich an, wenn man es geschafft hat. Du bist großartig.",
    "Träger des internationalen Ordens für kontrolliertes Chaos.",
    "Tschu-tschu! Der kleine Zug der Inkompetenz ist gerade noch rechtzeitig auf das richtige Gleis gefahren. Fein!",
    "Und der Oscar für die beste Darstellung eines Menschen, der angeblich alles unter Kontrolle hat, geht an: DICH!",
    "Und jetzt bitte alle aufstehen für diesen einmaligen Triumph der Improvisationskunst!",
    "Vergiss nicht, dir selbst heute etwas Gutes zu tun. Du hast es dir so was von verdient.",
    "Warum gehen Mahjong-Steine nicht zum Psychologen? Weil sie ihre Probleme lieber paarweise lösen.",
    "Warum gehen Mahjong-Steine nie alleine aus? Weil sie immer ein Gegenstück suchen!",
    "Warum hat der Mahjong-Spieler einen Knoten im Taschentuch? Damit er nicht vergisst, wo er den Stein gesehen hat!",
    "Warum hat der Mahjong-Spieler so gute Augen? Weil er ständig nach dem sucht, was direkt vor ihm liegt!",
    "Warum ist der letzte Stein immer so eingebildet? Weil alle auf ihn warten!",
    "Warum ist Mahjong das romantischste Spiel der Welt? Weil am Ende jeder seinen Zwilling findet.",
    "Warum ist Mahjong so romantisch? Am Ende finden alle ihr passendes Gegenstück!",
    "Warum können Mahjong-Steine nicht heiraten? Weil sie schon längst ein Paar sind!",
    "Warum können Mahjong-Steine nicht lügen? Weil ihnen alles ins Gesicht geschrieben steht.",
    "Warum können Mahjong-Steine nicht tanzen? Weil sie immer auf der gleichen Stelle stehen!",
    "Warum sind Mahjong-Spieler gute Detektive? Sie finden sogar Paare, die sich verstecken!",
    "Warum sind Mahjong-Spieler so geduldig? Weil sie wissen, dass irgendwann jeder Stein fällt!",
    "Warum sind Mahjong-Steine schlechte Geheimnisträger? Weil sie immer aufgedeckt werden.",
    "Warum sind Mahjong-Steine so beliebt? Sie haben einfach Charakter.",
    "Warum sind Mahjong-Steine so gute Freunde? Weil sie immer zusammenhalten!",
    "Warum sind Mahjong-Steine so ordentlich? Weil jeder seinen festen Platz hat. Meistens jedenfalls.",
    "Warum sind Mahjong-Steine so treu? Weil sie immer auf ihr Gegenstück warten!",
    "Warum sind Zwillinge beim Mahjong so beliebt? Weil sie immer zusammenpassen!",
    "Was ist das Lieblingslied eines Mahjong-Steins? „We Are Family!“",
    "Was ist das Schönste an Mahjong? Wenn am Ende alle wieder zusammenfinden.",
    "Was ist der Unterschied zwischen einem Mahjong-Spieler und einem Staubsauger? Der Staubsauger findet wenigstens alles!",
    "Was ist die größte Angst eines Mahjong-Steins? Für immer allein liegen zu bleiben!",
    "Was ist ein Mahjong-Stein ohne Partner? Ein trauriger Einzelfall!",
    "Was macht ein einsamer Mahjong-Stein? Er sucht seine bessere Hälfte – und hofft, dass sie nicht schon vergeben ist.",
    "Was macht ein Mahjong-Stein am Wochenende? Er hängt mit seinen Kumpels ab.",
    "Was macht ein Mahjong-Stein beim Friseur? Er lässt sich die Kanten schneiden.",
    "Was macht ein Mahjong-Stein beim Zahnarzt? Er lässt seine Ecken abschleifen.",
    "Was macht ein Mahjong-Stein im Urlaub? Er sucht seine bessere Hälfte!",
    "Was macht ein Mahjong-Stein, wenn er seinen Partner nicht findet? Er schaltet eine Kontaktanzeige: „Suche Zwilling, möglichst ohne Ecken und Kanten!“",
    "Was sagt der letzte Stein auf dem Spielfeld? „Ich bin dann mal weg!“",
    "Was sagt der letzte Stein zum Spieler? „Na, endlich hast du mich gefunden!“",
    "Was sagt der Stein zum Spiegel? „Heute sehe ich meinem Zwilling aber ähnlich!“",
    "Was sagt ein Mahjong-Spieler nach einem gewonnenen Level? „Das war doch ein Kinderspiel!“ Und dann kommt das nächste Level …",
    "Was sagt ein Mahjong-Stein zum anderen? „Wir passen einfach zusammen!“",
    "Was sagt ein Stein, wenn er endlich gefunden wird? „Ich dachte schon, du erkennst mich nicht mehr!“",
    "Weißt du eigentlich, wie schön es ist, dich über dich selbst hinauswachsen zu sehen?",
    "Weltmeister im professionellen Herumwurschteln mit anschließendem Spitzenergebnis.",
    "Wenn es für pure Genialität im Alleingang einen Preis gäbe, hättest du ihn soeben gestohlen.",
    "Wenn Sturheit eine olympische Disziplin wäre, hättest du gerade die Goldmedaille geholt.",
    "Wer braucht schon einen Doktortitel, wenn er Mahjong kann?",
    "Wer braucht schon einen Masterplan, wenn man pure, unbändige Willenskraft und etwas Galgenhumor hat?",
    "Wer braucht schon Logik, wenn man deinen ständigen Drang hat, alles auf den allerletzten Drücker zu regeln?",
    "Wer hätte gedacht, dass Aufräumen so viel Spaß machen kann?",
    "Wer hätte gedacht, dass aus dem totalen Chaos am Ende doch noch ein Meisterwerk wird?",
    "Wer hätte gedacht, dass in diesem Kopf so viel Effizienz und Skrupellosigkeit beim Lösen steckt?",
    "Wer ist denn hier der Retter in der Not? Fast wäre alles explodiert, aber du hast rechtzeitig den Stecker gezogen. Brav!",
    "Wie nennt man einen besonders schlauen Mahjong-Stein? Einen Hochstapler!",
    "Wieder ein Level geschafft. Dein Gehirn verlangt jetzt vermutlich Urlaub!",
    "Wieder ein Level geschafft. Langsam wirst du gefährlich!",
    "Wir hatten insgeheim schon die Rettungskräfte informiert, aber du hast es einfach so gelöst.",
    "Wissenschaftler nennen es Zufall. Wir nennen es Thomas-Effekt.",
    "Zwei gleiche Steine treffen sich. Sagt der eine: „Du kommst mir bekannt vor!“ Sagt der andere: „Kein Wunder, wir sehen uns zum Verwechseln ähnlich!“"
];

let gewonnenVorrat = [];
let letzterGewonnen = "";

function gewonnenSpruch() {
    if (gewonnenVorrat.length === 0) {
        gewonnenVorrat = GEWONNEN_SPRUECHE.slice();

        for (let i = gewonnenVorrat.length - 1; i > 0; i--) {
            const j = Math.floor(Math.random() * (i + 1));
            [gewonnenVorrat[i], gewonnenVorrat[j]] = [gewonnenVorrat[j], gewonnenVorrat[i]];
        }
    }

    let spruch = gewonnenVorrat.pop();

    if (spruch === letzterGewonnen && gewonnenVorrat.length > 0) {
        const tausch = gewonnenVorrat.pop();
        gewonnenVorrat.push(spruch);
        spruch = tausch;
    }

    letzterGewonnen = spruch;

    return spruch;
}

// Zeigt den Spruch unter der Siegmeldung
function zeigeGewonnenSpruch() {
    const feld = document.getElementById("gewonnenText");

    if (!feld) return;

    feld.textContent = gewonnenSpruch();
    feld.style.color = siegSpruchFarbe();
}

// Beim Start eines neuen Bretts wieder leeren
function leereGewonnenSpruch() {
    const feld = document.getElementById("gewonnenText");

    if (feld) feld.textContent = "";
}

// --------------------------------------------------
// SONDERZEICHEN ZWISCHEN SPRUCH UND WEITER-KNOPF
// --------------------------------------------------

const GEWONNEN_SYMBOLE = [
    "👑", "🏆", "🥇", "🎯", "☆", "💎", "✨", "✔", "⚡", "🔥",
    "❤", "💖", "👏", "👍", "💯", "🚀", "💥", "💫", "🔮"
];

let symbolVorrat = [];
let letztesSymbol = "";

function gewonnenSymbol() {
    if (symbolVorrat.length === 0) {
        symbolVorrat = GEWONNEN_SYMBOLE.slice();

        for (let i = symbolVorrat.length - 1; i > 0; i--) {
            const j = Math.floor(Math.random() * (i + 1));
            [symbolVorrat[i], symbolVorrat[j]] = [symbolVorrat[j], symbolVorrat[i]];
        }
    }

    let symbol = symbolVorrat.pop();

    if (symbol === letztesSymbol && symbolVorrat.length > 0) {
        const tausch = symbolVorrat.pop();
        symbolVorrat.push(symbol);
        symbol = tausch;
    }

    letztesSymbol = symbol;

    return symbol;
}

function zeigeGewonnenSymbol() {
    const feld = document.getElementById("gewonnenSymbol");

    if (feld) feld.textContent = gewonnenSymbol();
}

function leereGewonnenSymbol() {
    const feld = document.getElementById("gewonnenSymbol");

    if (feld) feld.textContent = "";
}
