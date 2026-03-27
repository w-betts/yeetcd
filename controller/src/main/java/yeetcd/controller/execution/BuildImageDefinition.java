package yeetcd.controller.execution;

import java.io.File;
import java.util.List;

public record BuildImageDefinition(String image, String tag, ImageBase imageBase, File artifactDirectory, List<String> artifiactNames, String cmd) {
}
