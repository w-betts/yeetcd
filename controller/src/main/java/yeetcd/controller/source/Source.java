package yeetcd.controller.source;

import org.apache.commons.codec.digest.DigestUtils;

public record Source(String name, byte[] zip) {

    public String sha256() {
        return DigestUtils.sha256Hex(zip);
    }
}
