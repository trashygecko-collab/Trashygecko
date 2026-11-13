/*
 * TRASHYGECKO MAHJONG
 * Sprueche fuer das geschaffte Level: erscheinen unter der Siegmeldung
 */

const GEWONNEN_SPRUECHE = [
    "Hätte absolut niemand drauf gewettet, aber du hast es tatsächlich durchgezogen.",
    "Ehrlich gesagt dachte ich, das endet im Desaster, aber chapeau, du hast abgeliefert.",
    "Die Statistik hat dagegen gesprochen, aber dein Sturkopf hat gewonnen.",
    "Niemand hat mit dir gerechnet – vor allem du selbst nicht –, aber du hast es geschafft.",
    "Das war haarscharf am Nervenzusammenbruch vorbei, aber das Ergebnis kann sich sehen lassen.",
    "Wir hatten insgeheim schon die Rettungskräfte informiert, aber du hast es einfach so gelöst.",
    "Wer hätte gedacht, dass aus dem totalen Chaos am Ende doch noch ein Meisterwerk wird?",
    "Du hast allen das Gegenteil bewiesen, die schon popcornbereit im Stuhl saßen.",
    "Selbst das Universum schien gegen dich zu arbeiten, aber du hast es trotzdem bezwungen.",
    "Ein modernes Wunder ist geschehen und du bist der alleinige Grund dafür.",
    "Alle sagten, das geht schief. Du hast es gemacht – und sie sprachlos zurückgelassen.",
    "Die Wahrscheinlichkeit sprach dagegen, aber wen kümmert schon Mathe, wenn man Talent hat?",
    "Ich war kurz davor, mein Testament zu machen, als du angefangen hast, aber du hast es tatsächlich überlebt.",
    "Niemand hat auch nur einen Cent auf dich gesetzt, und jetzt räumst du das Feld ab.",
    "Du hast den gordischen Knoten einfach durchgehackt, nachdem alle anderen daran verzweifelt sind.",
    "Es sah stundenlang nach Kernschmelze aus, aber du hast das Ding noch gerettet.",
    "Man soll den Tag nicht vor dem Abend loben, aber du hast ihn heute komplett rasiert.",
    "Die Erwartungen waren im Keller, aber du hast sie von unten noch untergraben – im positiven Sinne!",
    "Eigentlich war das absolut zum Scheitern verurteilt, aber du hast den Fluch gebrochen.",
    "Chapeau, ich dachte wirklich, das wird der Untergang, aber du hast es gerockt.",
    "Du hast das Ding überlebt und dabei sogar noch geschafft, gut auszusehen.",
    "Selbst nach dem gefühlten zehnten Herzinfarkt hast du es noch ins Ziel gebracht.",
    "Die Überlebenschancen waren gering, aber dein Ego hat es am Ende gerichtet.",
    "Man hätte dich für diesen waghalsigen Alleingang einweisen müssen, aber es hat funktioniert.",
    "Ein klassischer Fall von: Niemand wusste, wie es geht – außer dir, auch wenn du es erst im Nachhinein gemerkt hast.",
    "Du hast dem drohenden Untergang direkt ins Gesicht gelacht und gewonnen.",
    "Das war so nah an der Katastrophe, dass es schon wieder moderne Kunst war.",
    "Du hast das Unmögliche möglich gemacht, und jetzt müssen wir alle so tun, als wäre das normal.",
    "Selbst Murphys Gesetz hat vor deiner chaotischen Genialität kapituliert.",
    "Man stolpert manchmal über den Erfolg, aber du bist mit Anlauf reingesprungen und hast darin gebadet.",
    "Wenn Sturheit eine olympische Disziplin wäre, hättest du gerade die Goldmedaille geholt.",
    "Du hast den Karren nicht nur aus dem Dreck gezogen, sondern gleich noch ein Fluchtfahrzeug draus gemacht.",
    "Keine Ahnung, wie du das angestellt hast, ohne dass die Realität kollabiert ist, aber Respekt.",
    "Das hat mehr Nerven gekostet als eine Geiselnahme, aber das Ergebnis entschädigt für alles.",
    "Du hast das Schiff durch den Sturm manövriert, obwohl der Kapitän (du) absolut keinen Plan hatte.",
    "Wer braucht schon einen Masterplan, wenn man pure, unbändige Willenskraft und etwas Galgenhumor hat?",
    "Du hast den Karren so tief in den Sand gesetzt, dass du am anderen Ende der Welt wieder herausgekommen bist.",
    "Das war ein Ritt auf der Rasierklinge, aber du hast nicht mal einen Kratzer davongetragen.",
    "Manchmal muss man erst alles anzünden, um zu sehen, dass man es auch hätte retten können. Du hast es gerettet!",
    "Du hast den Berg erklommen, obwohl jeder wusste, dass es dort oben keinen Sauerstoff gibt.",
    "Selbst die Götter im Olymp staunen gerade neidisch über das, was du da abgeliefert hast.",
    "Wenn es für pure Genialität im Alleingang einen Preis gäbe, hättest du ihn soeben gestohlen.",
    "Man nennt es Talent, du nennst es Glück, ich nenne es illegale Hexerei.",
    "Du bist der lebende Beweis dafür, dass man mit genug Frechheit und Können jedes System sprengen kann.",
    "Ein Geniestreich, der extrem knapp an der legalen Grenze vorbeigeschrammt ist.",
    "Du hast das Ding so souverän gewuppt, als hättest du nie etwas anderes getan (außer Panik zu schieben).",
    "Die Geschichtsbücher werden diesen Moment ignorieren, aber wir wissen, was du da vollbracht hast.",
    "Du hast den ultimativen Bosskampf im Real Life einfach ohne Cheat-Codes durchgespielt.",
    "Einfach legendär, wie du alle Zweifel mit puren Fakten und Talent erschlagen hast.",
    "Wer hätte gedacht, dass in diesem Kopf so viel Effizienz und Skrupellosigkeit beim Lösen steckt?",
    "Du hast das System komplett ausgetrickst und bist jetzt der ungekrönte Star des Tages.",
    "Ein Meisterwerk der Improvisation, das während der Ausführung mehrfach kurz vor dem Einsturz stand.",
    "Du hast geliefert, als hätte dein Leben davon abgehängt – oder zumindest dein sehr stolzer Ruf.",
    "Das war so gut, dass es fast unfair gegenüber allen anderen ist, die es gar nicht erst versucht haben.",
    "Du hast das sprichwörtliche Brett vorm Kopf abgenommen und direkt damit das Tor eingeschlagen.",
    "Respekt, du hast das Unmögliche geschafft und dabei nicht mal eine Miene verzogen.",
    "Du bist wie ein Phoenix aus der Aschenbahn der totalen Inkompetenz aufgestiegen.",
    "Das war der perfekte Beweis dafür, dass man mit genügend Chaos immer ans Ziel kommt.",
    "Du hast die Messlatte so unfassbar hoch gelegt, dass wir sie wohl nie wieder finden werden.",
    "Ein monumentaler Sieg gegen jegliche Gesetze des gesunden Menschenverstandes.",
    "Du hast es geschafft, ohne dass wir uns seelisch auch nur ansatzweise darauf einstellen konnten.",
    "Das war knapp, aber knapp ist bekanntlich auch vorbei – und vorbei ist gewonnen!",
    "Du hast das Unmögliche geschafft, und jetzt tun wir alle so, als wäre das eine Kleinigkeit gewesen.",
    "Wer braucht schon Logik, wenn man deinen ständigen Drang hat, alles auf den allerletzten Drücker zu regeln?",
    "Du hast es durchgezogen wie ein eiskalter Profi, der heimlich im Keller geübt hat.",
    "Das war so brillant und dreist, dass es fast schon wieder verdächtig ist.",
    "Du hast den Sack zugemacht, bevor überhaupt jemand gemerkt hat, dass die Katastrophe entwischt war.",
    "Ein historischer Triumph der reinen Willenskraft über den gesunden Verstand.",
    "Du hast das Ding gerockt, während der Rest der Welt noch panisch die Anleitung gesucht hat.",
    "Man muss dir eins lassen: Wenn du etwas an die Wand fährst, baust du wenigstens direkt ein Kunstwerk daraus.",
    "Du hast den Ernst der Lage erkannt und sie prompt ignoriert – mit vollem, glorreichen Erfolg.",
    "Das war ein Husarenstück, das selbst Baron von Münchhausen vor Neid erblassen lassen würde.",
    "Du hast das Ruder im allerletzten Moment herumgerissen, als das Schiff schon lichterloh brannte.",
    "Hut ab, du hast das Ding nach Hause gebracht, obwohl der Kompass längst geschmolzen war.",
    "Du hast das Unmögliche möglich gemacht und siehst dabei noch so aus, als ob es ein öder Mittwoch wäre.",
    "Ein denkwürdiger Moment: Du hast etwas Ernstes angefangen und es tatsächlich zu einem Ende gebracht.",
    "Du hast den gordischen Knoten nicht nur gelöst, sondern ihn direkt feierlich verbrannt.",
    "Das war der ultimative Beweis: Wer keine Ahnung hat, improvisiert sich einfach souverän an die Spitze.",
    "Du hast es geschafft, und das Schlimmste daran ist: Es sieht auch noch verdammt gut aus.",
    "Du hast den Hauptgewinn geknackt, obwohl du anfangs nicht mal wusstest, wo der Automat steht.",
    "Du hast den Kopf im letzten Moment aus der Schlinge gezogen und den Strick direkt als stylische Krawatte umgebunden.",
    "Das war so souverän, dass selbst die hartnäckigsten Skeptiker ihre Popcorn-Tüten weggelegt haben.",
    "Du hast das Schlachtfeld betreten, alles niedergerissen und stehst jetzt auf den Trümmern – erfolgreich!",
    "Man fragt sich manchmal, ob du ein verkanntes Genie oder einfach nur lebensmüde bist, aber das Ergebnis stimmt.",
    "Du hast dem finsteren Abgrund direkt angelächelt und bist einfach drübergesprintet.",
    "Ein Triumph sondergleichen, der dich vermutlich mindestens drei Jahre Lebenszeit gekostet hat.",
    "Du hast die Mission erfüllt, obwohl die Zielscheibe die ganze Zeit auf deinem eigenen Rücken klebte.",
    "Das war kein einfacher Schritt nach vorn, das war ein gewaltiger Sprung über den eigenen Schatten.",
    "Du hast den Nagel so präzise auf den Kopf getroffen, dass der Hammer in tausend Teile zersprungen ist.",
    "Du hast das Unmögliche geschafft, und jetzt verneigen sich selbst die Kritiker, die schon mit dem Grabstein wedelten.",
    "Das war der endgültige Beweis, dass man mit einer gesunden Portion Wahnsinn im Leben alles erreichen kann.",
    "Du hast das Ding so genial gedreht, dass die Physik kurz überlegt hat, ihre eigenen Gesetze zu ändern.",
    "Hut ab, du hast aus dem drohenden Totalausfall einen absoluten Blockbuster gemacht.",
    "Du hast den Berg erklommen, den alle anderen schon längst weggesprengt hatten.",
    "Ein grandioser Sieg, der exakt so geplant war – zumindest in deinen kühnsten Albträumen.",
    "Du hast es geschafft, und das Beste daran ist: Niemand kann jetzt behaupten, du hättest auch nur einen Hauch von Hilfe gehabt.",
    "Du hast das Unmögliche geschafft, und jetzt trinken wir kollektiv auf deine absolut schamlose Sturheit.",
    "Das war ein Husarenstreich, der so haarscharf an einer Kernkatastrophe vorbeigeschrammt ist, dass es fasziniert.",
    "Du hast das Ziel erreicht, obwohl der Pfad seit Jahren von Dornen und Verzweiflung überwuchert war.",
    "Du hast es geschafft – und jetzt atmen wir alle tief aus und trinken einen auf deine Nerven aus Stahl.",
    "a, wer ist denn hier das ganz große Genie? Hätte zwar fast die Realität gesprengt, aber hat geklappt! Ja, feini!",
    "Ei, guck mal an! Ganz alleine den Karren in den Dreck gefahren und im letzten Moment wieder rausgezogen. Wer ist ein kluges Köpfchen? Du!",
    "Fein gemacht! Fast hätten wir den Notarzt gebraucht, aber du hast den richtigen Knopf gedrückt. Tschakka!",
    "Ja, huch! Wer hat denn da das unlösbare Problem gelöst, ohne dass das Universum implodiert ist? Du feiner Spatz!",
    "Guck mal, der kleine Wirbelwind hat den Weltuntergang einfach mal um eine Woche verschoben. Ei, wie süß!",
    "Ei, ei, ei – fast hätt's gekracht, aber du hast den Kopf aus der Schlinge gezogen. Wer kriegt jetzt ein Gummibärchen? Du!",
    "Da hat der kleine Strolch doch tatsächlich das Chaos bezwungen, bevor die Erwachsenen panisch die Evakuierung starten mussten. Fein!",
    "Ja, wer kann denn da so toll Tasten drücken, dass die Katastrophe ausbleibt? Ja, du bist es!",
    "Ei, wer hat denn das schwere Ding ganz alleine geschafft, ohne dass echte Tränen geflossen sind? (Fast-Tränen zählen nicht!)",
    "Hui! Da staunt der Laie und der Fachmann weint – aber du hast es geschafft, mein kleiner Chaot!",
    "Ja, feini gemacht! Wer hat den gordischen Knoten durchgehackt und dabei so getan, als wüsste man, was man tut? Du!",
    "Ei, schau mal einer an: Vom Totalausfall zum Überlebenskünstler in drei Sekunden. Wer ist ein braves Kind?",
    "Tschu-tschu! Der kleine Zug der Inkompetenz ist gerade noch rechtzeitig auf das richtige Gleis gefahren. Fein!",
    "Ja, huch, es hat geklappt! Ohne dass wir das Gebäude räumen mussten. Da machen wir drei Kreuzchen!",
    "Ei, wer hat denn dem drohenden Untergang so tapfer in die Augen geblickt? Du kleiner Draufgänger!",
    "Wer ist denn hier der Retter in der Not? Fast wäre alles explodiert, aber du hast rechtzeitig den Stecker gezogen. Brav!",
    "Ja, ei, ei, da guck her: Ein echtes kleines Wunder, weil die Wahrscheinlichkeit für diesen Erfolg bei exakt null lag. Toll gemacht!",
    "Hui, das war knapp an der Kernschmelze vorbei! Aber wer hat's gerettet? Unser kleiner Problemlöser!",
    "Ja, fein! Wer hat den Karren aus dem Sumpf gezogen und sich dabei nur minimal dreckig gemacht? Du!",
    "Ei, schau mal, die Welt existiert noch, weil du im letzten Moment den Daumen draufgehalten hast. Feines Kind!",
    "Ja, wer hat denn da so schön improvisiert, dass selbst die Physik kurz kapituliert hat? Du kleiner Magier!",
    "Ei, ei, ei, das war haarscharf vor dem Nervenzusammenbruch, aber jetzt wird geklatscht! Fein gemacht!",
    "Ja, wer ist denn der Champion des Hauses nach diesem heroischen Überlebenskampf? Du!",
    "Guck an, der kleine Pechvogel hat sich einfach zum Glückspilz hochgestolpert. Ei, wie niedlich!",
    "Ja, fein gemacht! Erst alles anzünden und dann geschickt löschen. Wer ist ein kluges Köpfchen?",
    "Ei, wer hat den Fehler behoben, ohne dass das System kollabierte? Ja, du!",
    "Hui, da klopfen wir uns aber mal auf die Schulter! Der kleine Überlebenskünstler hat's mal wieder geschafft.",
    "Ja, ei, wer hat denn da allen gezeigt, dass man mit purer Sturheit jede Statistik schlägt? Du kleiner Sturkopf!",
    "Fein, fein! Fast wäre es ein Desaster geworden, aber du hast die Kurve noch gekratzt. Brav!",
    "Ei, wer kriegt jetzt ein Fleißbändchen für das Verhindern des totalen Ruins? Du natürlich!",
    "Ja, huch! Da staunt man aber, dass das tatsächlich funktioniert hat. Wer ist ein schlaues Köpfchen?",
    "Ei, ei, ei, wer hat den fetten Patzer einfach durch freches Weitermachen ausgebügelt? Du kleiner Rebell!",
    "Ja, fein! Der Abgrund hat gerufen, aber du hast ihm einfach die Zunge rausgestreckt und gewonnen.",
    "Hui, wer hat denn da den Mega-Fehler ausgebügelt, bevor es jemand bemerkt hat? Unser kleiner Ninja!",
    "Ei, schau mal einer an, der kleine Tollpatsch ist im Ziel angekommen und stolpert direkt über den Siegerpokal!",
    "Ja, feini! Fast in die Hose gegangen, aber am Ende glanzvoll gerettet. Wer ist der Beste? Du!",
    "Ei, wer hat den Karren so tief reingefahren und ihn dann doch noch rausgezogen? Eine Meisterleistung der Verwirrung!",
    "Ja, huch, es ist vollbracht! Und niemand musste weinen (außer der Computer). Brav!",
    "Ei, schau an, der kleine Überlebenskünstler hat den Bosskampf im Real Life ohne Cheat geschafft. Fein!",
    "Ja, wer hat denn da so toll den Kopf aus der Schlinge gezogen? Dafür gibt's ein virtuelles Kügelchen Eis!",
    "Hui, da hat aber jemand gezittert! Aber das Ergebnis zählt. Wer ist ein tapferes Kind?",
    "Ei, ei, ei, wer hat das Unmögliche möglich gemacht und sieht dabei noch völlig unschuldig aus? Du!",
    "Ja, feini gemacht! Der Untergang war nah, aber deine Ignoranz war stärker. Glückwunsch!",
    "Ei, schau mal, der kleine Chaos-König hat den Laden gerettet. Wer hätte das gedacht?",
    "Ja, wer hat denn da den gordischen Knoten zerfetzt und sich gefreut wie ein Schneekönig? Du!",
    "Hui, das war ein Ritt auf der Rasierklinge, aber du hast nicht mal einen Kratzer. Brav!",
    "Ei, wer ist denn hier der ungeschlagene Champion im Haarscharf-am-Desaster-Vorbeischrammen? Du!",
    "Ja, fein gemacht! Das hätte auch sehr, sehr böse enden können, aber Glück muss der Mensch haben.",
    "Ei, schau an, der kleine Meister des Notfalls hat wieder zugeschlagen. Wer ist ein Genie? Du!",
    "Ja, huch! Da hat der kleine Strolch tatsächlich geliefert, als es drauf ankam. Feini, feini!",
    "Die Kommission zur Feststellung historischer Großtaten hat getagt: Deine Bewältigung dieses kolossalen Desasters wird als unumstößlicher Meilenstein in die Annalen des Überlebens eingehen.",
    "Angesichts der absoluten Unvermeidbarkeit des Scheiterns hast du mit unerbittlicher, kalter Entschlossenheit die Realität selbst gebogen. Wir verneigen uns in eisiger Ehrfurcht.",
    "Die Naturgesetze schrien in Agonie, die Statistik forderte deinen sofortigen Rücktritt – und dennoch steht das Werk. Dies ist kein bloßer Erfolg, das ist ein Triumph über die Logik.",
    "Die Würde des Augenblicks verbietet jeden billigen Jubel. Stattdessen gebührt dir das schweigende, unumstößliche Anerkennen einer monumentalen, fast schon beängstigenden Kompetenz.",
    "Man hat schon viele Helden fallen sehen, doch wer aus dem Auge des absoluten Sturms mit einem funktionierenden Ergebnis hervorgeht, verdient keinen Applaus, sondern ein Denkmal aus kaltem Marmor.",
    "Die Zivilisation stand am Rande des Abgrunds, bereit, sich selbst aufzugeben. Du hast nicht nur eingegriffen, du hast das Schicksal eigenhändig zur Raison gebracht.",
    "Atme erst mal tief aus, mein Schatz. Das Schlimmste ist vorbei, und du hast das so unfassbar gut gemacht.",
    "Komm erst mal an, lehn dich zurück und lass den ganzen Druck von dir abfallen. Du hast es geschafft.",
    "Sieh mich an: Du hast es hinter dir und es ist gut ausgegangen. Jetzt darfst du dich einfach nur ausruhen.",
    "Alles ist gut. Du hast so viel Energie reingesteckt, jetzt darfst du dich einfach mal zurücklehnen und stolz auf dich sein.",
    "Spürst du, wie die Anspannung abfällt? Genau so fühlt sich an, wenn man es geschafft hat. Du bist großartig.",
    "Ich bin einfach so unglaublich stolz auf dich – nicht nur, weil es geklappt hat, sondern weil ich weiß, wie viel Kraft dich das gekostet hat.",
    "Egal, was alle anderen sagen: Ich habe gesehen, wie hart du dafür gekämpft hast. Und das Ergebnis ist einfach wunderbar.",
    "Du kannst so verdammt stolz auf dich sein. Ich bin es auf jeden Fall, und zwar von ganzem Herzen.",
    "Niemand muss dir mehr beweisen, was in dir steckt, du hast es dir selbst und allen anderen gerade eindrucksvoll gezeigt.",
    "Ich wusste zu jeder Sekunde, dass du das schaffst – auch dann, als du selbst nicht mehr daran geglaubt hast.",
    "Es war ein schwerer Weg und du hast gezweifelt, aber du bist durchgegangen. Das zeigt einfach, wie stark du bist.",
    "Du hast dir so viele Sorgen gemacht und nachts gegrübelt – und sieh nur, wie toll du das am Ende gelöst hast.",
    "Du darfst jetzt weich werden und nachgeben. Die Anspannung war riesig, aber du hast es getragen. Ganz allein.",
    "Es ist völlig normal, dass du jetzt erschöpft bist. Du hast Großes geleistet und darfst dich jetzt feiern lassen.",
    "Du hast so oft an dir gezweifelt, aber deine Ausdauer hat gewonnen. Du bist einfach bewundernswert.",
    "Du bist wunderbar, genau so wie du bist. Und heute hast du der Welt (und dir selbst) wieder gezeigt, was für ein toller Mensch in dir steckt.",
    "Ich drücke dich in Gedanken ganz fest. Du hast das so sauber und mit so viel Herz durchgezogen.",
    "Vergiss nicht, dir selbst heute etwas Gutes zu tun. Du hast es dir so was von verdient.",
    "Du bist ein echtes Geschenk, und ich freue mich einfach riesig mit dir über diesen Erfolg.",
    "Mach dir jetzt einen schönen Abend, genieß diesen Moment und lass dir gesagt sein: Du bist einfach spitze, und ich bin unendlich froh, dass es dich gibt.",
    "Gewinner des goldenen Schraubenschlüssels für außergewöhnliche Improvisation.",
    "Träger des internationalen Ordens für kontrolliertes Chaos.",
    "Ehrenmitglied der Gesellschaft zur Verhinderung unnötiger Katastrophen.",
    "Weltmeister im professionellen Herumwurschteln mit anschließendem Spitzenergebnis.",
    "Ausgezeichnet mit dem silbernen Rettungsring für außergewöhnliche Selbstrettung.",
    "Gewinner des Preises für die kreativste Umgehung des gesunden Menschenverstands.",
    "Nicht der Weg macht den Sieger aus, sondern der Mut, ihn trotz aller Zweifel zu gehen.",
    "Große Leistungen beginnen oft dort, wo andere längst aufgegeben haben.",
    "Manchmal ist Durchhalten die größte Form von Intelligenz.",
    "Nicht jeder Triumph ist laut. Manche bestehen einfach darin, nicht aufgegeben zu haben.",
    "Du hast heute nicht nur ein Problem gelöst, sondern auch ein Stück Vertrauen in dich selbst zurückgewonnen.",
    "Ach, du kleines Wunder auf zwei Beinen. Wer hätte dir das zugetraut?",
    "Ich könnte dich gerade gleichzeitig knuddeln und für deinen Lösungsweg ausschimpfen.",
    "Du bist manchmal eine Herausforderung für die Menschheit, aber heute eine verdammt erfolgreiche.",
    "Komm her, du verrückter Schatz. Das hast du wirklich toll gemacht.",
    "Ein bisschen bekloppt, aber dafür außergewöhnlich erfolgreich.",
    "Ich wusste doch, dass unter dem ganzen Chaos ein kleines Genie wohnt.",
    "Eilmeldung: Ein Mensch hat tatsächlich etwas zu Ende gebracht. Die Bevölkerung jubelt.",
    "Sensationelle Szenen spielen sich ab. Experten sprechen von einem historischen Durchbruch.",
    "Die Börse reagiert positiv, nachdem dein Erfolg offiziell bestätigt wurde.",
    "Hiermit wird Ihnen die uneingeschränkte Befähigung zur Problembewältigung bescheinigt.",
    "Nach eingehender Prüfung wurde festgestellt: Das war tatsächlich nicht schlecht.",
    "Ihr Antrag auf Anerkennung Ihrer Genialität wurde einstimmig genehmigt.",
    "Die zuständige Behörde ist sprachlos und hat den Vorgang geschlossen.",
    "Macht Platz! Hier kommt jemand, der sogar dem Zufall Anweisungen gibt.",
    "Ab heute wird dein Name in Großbuchstaben geschrieben.",
    "Ich möchte nicht übertreiben, aber die NASA hat gerade nach deinem Lösungsweg gefragt.",
    "Das war dermaßen unerwartet, dass ich kurz meine Berechnungen überprüfen musste.",
    "Ich hätte dir ja geholfen, aber offenbar war meine Anwesenheit völlig überflüssig.",
    "Du bist der Grund, warum Bedienungsanleitungen inzwischen Warnhinweise haben.",
    "Du hast zwar keine Ahnung gehabt, aber dafür eine beeindruckende Konsequenz an den Tag gelegt.",
    "Das war nicht unbedingt intelligent, aber verdammt effektiv.",
    "Erhebt euch! Seine Majestät der Problemlösung hat gesprochen.",
    "Ein dreifaches Hoch auf den Bezwinger der völlig unnötigen Katastrophen!",
    "Ab sofort darfst du deine eigenen Gesetze schreiben. Die alten scheinen dich ohnehin nicht aufzuhalten.",
    "Eine neue Studie bestätigt: Du bist zu 73 Prozent Genie und zu 27 Prozent unberechenbares Risiko.",
    "Du bist manchmal ein wandelnder Systemfehler, aber heute einer mit hervorragendem Ergebnis.",
    "Weißt du eigentlich, wie schön es ist, dich über dich selbst hinauswachsen zu sehen?",
    "Du musst nicht immer stark sein. Aber heute darfst du ruhig stolz auf deine Stärke sein.",
    "Du schaffst das! Und wenn nicht, behaupten wir einfach, es war ein Test.",
    "Kopf hoch, Brust raus, Bauch rein – und jetzt tu so, als hättest du einen Plan!",
    "Du bist nicht gescheitert. Du hast lediglich eine alternative Erfolgsroute entdeckt.",
    "Aufgeben ist keine Option. Höchstens eine kurze Kaffeepause mit dramatischem Blick aus dem Fenster.",
    "Denk immer daran: Irgendwo gibt es jemanden, der noch weniger Ahnung hat als du. Und der hat heute frei.",
    "Ich hoffe, du siehst dich für einen Moment mit meinen Augen. Dann würdest du verstehen, wie viel du wert bist.",
    "Auch kleine Schritte können riesige Siege sein. Und deiner zählt heute ganz besonders.",
    "Du bist nicht nur dann wunderbar, wenn dir alles gelingt. Du bist es auch an den Tagen, an denen du einfach nur weitermachst.",
    "Ich würde sagen, du hast einen an der Waffel. Aber offenbar ist genau das dein Erfolgsgeheimnis.",
    "Dein innerer Kompass zeigt zwar gelegentlich Richtung Frittenbude, aber heute hat er den richtigen Weg gefunden.",
    "Du bist eine Zumutung für die Logik und eine Bereicherung für die Erfolgsgeschichte.",
    "Bei dir weiß man nie, ob man helfen, lachen oder vorsichtshalber die Feuerwehr rufen soll.",
    "Forscher haben dein Vorgehen untersucht und anschließend beschlossen, lieber etwas anderes zu studieren.",
    "Und der Oscar für die beste Darstellung eines Menschen, der angeblich alles unter Kontrolle hat, geht an: DICH!",
    "Eine Leistung, die selbst Hollywood nur mit Spezialeffekten glaubwürdig darstellen könnte.",
    "Die Jury war sich uneinig, ob sie applaudieren oder einen Krankenwagen bestellen sollte.",
    "Und jetzt bitte alle aufstehen für diesen einmaligen Triumph der Improvisationskunst!",
    "Das war kein gewöhnlicher Erfolg. Das war ein dreistündiger Blockbuster mit Happy End.",
    "Dein Erfolg widerspricht mindestens drei wissenschaftlichen Theorien und einer Bedienungsanleitung.",
    "Die Wahrscheinlichkeit deines Erfolgs wurde gerade wegen mangelnder Beweise aus der Statistik entfernt.",
    "Wissenschaftler nennen es Zufall. Wir nennen es Thomas-Effekt.",
    "Dein Gehirn arbeitet offenbar mit einer Software, für die es noch keine Updates gibt.",
    "Dein königlicher Titel lautet: Herrscher über Chaos, Zufall und gelegentliche Geistesblitze.",
    "Die Krone sitzt. Der Verstand ist optional. Der Erfolg ist amtlich.",
    "Man möge dir Weintrauben reichen und sämtliche Aufgaben von dir fernhalten.",
    "Andere brauchen einen Plan. Du brauchst offenbar nur ausreichend Selbstvertrauen.",
    "Ich möchte deine Vorgehensweise nicht hinterfragen. Ich möchte sie einfach nur überleben.",
    "Du hast das Problem nicht verstanden, aber ihm trotzdem erfolgreich den Garaus gemacht.",
    "Das war fachlich vielleicht fragwürdig, menschlich aber eine Glanzleistung.",
    "Du bist der lebende Beweis, dass man auch mit einer Bedienungsanleitung auf dem Kopf erfolgreich sein kann.",
    "So viel Glück kann man eigentlich gar nicht haben. Du hast es trotzdem geschafft.",
    "Ich würde sagen, das war geplant. Wir müssen ja nicht alles verraten.",
    "Selbst dein innerer Schweinehund hat gerade fristlos gekündigt.",
    "Ich bin beeindruckt. Und das passiert mir wirklich nicht oft.",
    "Das Ergebnis ist gut. Über den Weg dorthin reden wir ein anderes Mal.",
    "Die Welt hat einen neuen Maßstab für außergewöhnliche Leistungen.",
    "Bitte einmal kurz aufstehen. Eine Legende hat gerade den Raum betreten.",
    "Bescheidenheit ist heute leider ausverkauft.",
    "Dein Denkmal wird selbstverständlich aus purem Gold gefertigt.",
    "Für diese Leistung werden Ihnen rückwirkend sämtliche Zweifel erlassen.",
    "Der Ausschuss empfiehlt dringend eine angemessene Selbstbewunderung.",
    "Ihr persönlicher Erfolgsstatus wurde auf „nicht mehr zu toppen“ hochgestuft.",
    "Sondersendung: Du hast geliefert. Weitere Informationen liegen noch nicht vor.",
    "Augenzeugen berichten von ungläubigem Staunen und spontanen Standing Ovations."
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

    if (feld) feld.textContent = gewonnenSpruch();
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
