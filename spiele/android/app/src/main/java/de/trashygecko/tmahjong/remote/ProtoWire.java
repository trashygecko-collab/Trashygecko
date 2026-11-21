package de.trashygecko.tmahjong.remote;

import java.io.ByteArrayOutputStream;
import java.io.EOFException;
import java.io.IOException;
import java.io.InputStream;
import java.io.OutputStream;
import java.nio.charset.StandardCharsets;
import java.util.ArrayList;
import java.util.List;

/** Small bounded protobuf wire reader/writer for the documented message subset used by Android TV Remote v2. */
final class ProtoWire {
    static final int MAX_FRAME = 256 * 1024;

    static final class Field {
        final int number;
        final int wireType;
        final long integer;
        final byte[] data;
        Field(int number, int wireType, long integer, byte[] data) {
            this.number = number; this.wireType = wireType; this.integer = integer; this.data = data;
        }
        String string() { return data == null ? "" : new String(data, StandardCharsets.UTF_8); }
    }

    private ProtoWire() { }

    static byte[] varintField(int number, long value) {
        ByteArrayOutputStream out = new ByteArrayOutputStream();
        writeVarint(out, ((long) number << 3)); writeVarint(out, value); return out.toByteArray();
    }
    static byte[] bytesField(int number, byte[] value) {
        ByteArrayOutputStream out = new ByteArrayOutputStream();
        writeVarint(out, ((long) number << 3) | 2); writeVarint(out, value.length); out.write(value, 0, value.length); return out.toByteArray();
    }
    static byte[] stringField(int number, String value) { return bytesField(number, value.getBytes(StandardCharsets.UTF_8)); }
    static byte[] join(byte[]... parts) {
        ByteArrayOutputStream out = new ByteArrayOutputStream();
        for (byte[] part : parts) if (part != null) out.write(part, 0, part.length);
        return out.toByteArray();
    }
    static byte[] outer(int fieldNumber, byte[] inner) {
        return join(varintField(1, 2), varintField(2, 200), bytesField(fieldNumber, inner));
    }
    static byte[] remote(int fieldNumber, byte[] inner) { return bytesField(fieldNumber, inner); }

    static List<Field> parse(byte[] input) throws IOException {
        if (input == null || input.length > MAX_FRAME) throw new IOException("Protobuf-Nachricht ist ungültig oder zu groß.");
        List<Field> fields = new ArrayList<>();
        int[] cursor = {0};
        while (cursor[0] < input.length) {
            long tag = readVarint(input, cursor);
            int number = (int) (tag >>> 3), wire = (int) (tag & 7);
            if (number <= 0) throw new IOException("Ungültiges Protobuf-Feld.");
            if (wire == 0) fields.add(new Field(number, wire, readVarint(input, cursor), null));
            else if (wire == 2) {
                long length = readVarint(input, cursor);
                if (length < 0 || length > MAX_FRAME || cursor[0] + length > input.length) throw new IOException("Ungültige Protobuf-Länge.");
                int end = cursor[0] + (int) length;
                byte[] value = new byte[(int) length]; System.arraycopy(input, cursor[0], value, 0, value.length); cursor[0] = end;
                fields.add(new Field(number, wire, 0, value));
            } else if (wire == 1 || wire == 5) {
                int skip = wire == 1 ? 8 : 4;
                if (cursor[0] + skip > input.length) throw new EOFException("Unvollständiges Protobuf-Feld.");
                cursor[0] += skip;
            } else throw new IOException("Nicht unterstützter Protobuf-Wire-Typ.");
        }
        return fields;
    }
    static Field first(List<Field> fields, int number) {
        for (Field field : fields) if (field.number == number) return field;
        return null;
    }
    static List<Field> all(List<Field> fields, int number) {
        List<Field> found = new ArrayList<>();
        for (Field field : fields) if (field.number == number) found.add(field);
        return found;
    }
    static byte[] readDelimited(InputStream in) throws IOException {
        long size = readVarint(in);
        if (size <= 0 || size > MAX_FRAME) throw new IOException("Ungültige Länge des TV-Protokollrahmens.");
        byte[] data = new byte[(int) size]; int offset = 0;
        while (offset < data.length) { int count = in.read(data, offset, data.length - offset); if (count < 0) throw new EOFException("TV-Verbindung wurde während einer Antwort geschlossen."); offset += count; }
        return data;
    }
    static void writeDelimited(OutputStream out, byte[] payload) throws IOException {
        if (payload == null || payload.length == 0 || payload.length > MAX_FRAME) throw new IOException("Ungültige Protokollnachricht.");
        ByteArrayOutputStream prefix = new ByteArrayOutputStream(); writeVarint(prefix, payload.length); out.write(prefix.toByteArray()); out.write(payload); out.flush();
    }
    static void writeVarint(ByteArrayOutputStream out, long value) {
        while ((value & ~0x7fL) != 0) { out.write((int) ((value & 0x7f) | 0x80)); value >>>= 7; }
        out.write((int) value);
    }
    private static long readVarint(byte[] input, int[] cursor) throws IOException {
        long value = 0; int shift = 0;
        while (shift < 64 && cursor[0] < input.length) {
            int b = input[cursor[0]++] & 0xff; value |= (long) (b & 0x7f) << shift;
            if ((b & 0x80) == 0) return value;
            shift += 7;
        }
        throw new IOException("Ungültiges oder unvollständiges Protobuf-Varint.");
    }
    private static long readVarint(InputStream input) throws IOException {
        long value = 0; int shift = 0;
        while (shift < 64) { int b = input.read(); if (b < 0) throw new EOFException("TV hat keinen Protokollrahmen gesendet."); value |= (long) (b & 0x7f) << shift; if ((b & 0x80) == 0) return value; shift += 7; }
        throw new IOException("Ungültige Protobuf-Rahmenlänge.");
    }
}