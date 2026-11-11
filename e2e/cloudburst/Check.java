// Check decodes a packet dump of the e2e client (-dump) with CloudburstMC's
// codec for that protocol: an independent implementation, closer to what real
// clients accept than gophertunnel, which reads some fields more leniently.
//
//   java -cp <cloudburst classpath> Check.java <protocol> <dump file> [game version]
//
// 1.19.62 kept protocol 567 but took 1.19.63's format (568), as in Geyser.
//
// Prints one line per packet type that fails to decode or leaves bytes over,
// then "OK" or "FAIL".
import io.netty.buffer.ByteBuf;
import io.netty.buffer.Unpooled;
import org.cloudburstmc.protocol.bedrock.codec.BedrockCodec;
import org.cloudburstmc.protocol.bedrock.codec.BedrockCodecHelper;
import org.cloudburstmc.protocol.bedrock.data.EncodingSettings;
import org.cloudburstmc.protocol.bedrock.data.PacketRecipient;
import org.cloudburstmc.protocol.bedrock.data.definitions.BlockDefinition;
import org.cloudburstmc.protocol.bedrock.data.definitions.ItemDefinition;
import org.cloudburstmc.protocol.bedrock.data.definitions.SimpleItemDefinition;
import org.cloudburstmc.protocol.bedrock.data.definitions.SimpleNamedDefinition;
import org.cloudburstmc.protocol.common.DefinitionRegistry;
import org.cloudburstmc.protocol.common.NamedDefinition;

import java.nio.file.Files;
import java.nio.file.Path;
import java.util.HexFormat;
import java.util.Map;
import java.util.TreeMap;

public class Check {
    record Block(int id) implements BlockDefinition {
        public int getRuntimeId() { return id; }
    }

    public static void main(String[] args) throws Exception {
        int proto = Integer.parseInt(args[0]);
        int codecProto = proto == 567 && args.length > 2 && !args[2].matches("1\\.19\\.6[01]") ? 568 : proto;
        BedrockCodec codec = (BedrockCodec) Class.forName("org.cloudburstmc.protocol.bedrock.codec.v" + codecProto + ".Bedrock_v" + codecProto)
                .getField("CODEC").get(null);
        BedrockCodecHelper helper = codec.createHelper();
        helper.setEncodingSettings(EncodingSettings.CLIENT); // the limits of a client, not of a server
        // Every runtime ID is accepted: this checks the wire format, not the registries.
        helper.setItemDefinitions(new DefinitionRegistry<ItemDefinition>() {
            public ItemDefinition getDefinition(int id) { return new SimpleItemDefinition("unknown:" + id, id, false); }
            public ItemDefinition getDefinition(String name) { return new SimpleItemDefinition(name, 0, false); }
            public boolean isRegistered(ItemDefinition d) { return true; }
        });
        helper.setBlockDefinitions(new DefinitionRegistry<BlockDefinition>() {
            public BlockDefinition getDefinition(int id) { return new Block(id); }
            public boolean isRegistered(BlockDefinition d) { return true; }
        });
        try { // only the codecs with camera presets (1.20.30+) take a registry
            helper.setCameraPresetDefinitions(new DefinitionRegistry<NamedDefinition>() {
                public NamedDefinition getDefinition(int id) { return new SimpleNamedDefinition("preset:" + id, id); }
                public boolean isRegistered(NamedDefinition d) { return true; }
            });
        } catch (UnsupportedOperationException ignored) {
        }

        Map<String, Integer> ok = new TreeMap<>();
        Map<String, String> bad = new TreeMap<>();
        for (String line : Files.readAllLines(Path.of(args[1]))) {
            if (line.isBlank()) continue;
            String[] f = line.split(" ", 2);
            int id = Integer.parseInt(f[0]);
            ByteBuf buf = Unpooled.wrappedBuffer(HexFormat.of().parseHex(f.length > 1 ? f[1] : ""));
            String name = "id " + id;
            try {
                var def = codec.getPacketDefinition(id);
                if (def == null) { bad.putIfAbsent(name, "not a packet of this protocol"); continue; }
                name = def.getFactory().get().getClass().getSimpleName();
                Object pk = codec.tryDecode(helper, buf, id, PacketRecipient.CLIENT);
                if (buf.isReadable()) bad.putIfAbsent(name, buf.readableBytes() + " bytes left over after " + pk);
                else ok.merge(name, 1, Integer::sum);
            } catch (Throwable t) {
                Throwable c = t; while (c.getCause() != null) c = c.getCause();
                bad.putIfAbsent(name, c.toString());
            }
        }
        bad.forEach((k, v) -> System.out.println("  " + k + ": " + (v.length() > 300 ? v.substring(0, 300) + "…" : v)));
        System.out.println((bad.isEmpty() ? "OK" : "FAIL") + " " + ok.values().stream().mapToInt(i -> i).sum() + " packets decoded, " + ok.size() + " types" + (bad.isEmpty() ? "" : ", " + bad.size() + " types failed"));
        System.exit(bad.isEmpty() ? 0 : 1);
    }
}
