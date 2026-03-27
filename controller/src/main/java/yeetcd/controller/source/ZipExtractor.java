package yeetcd.controller.source;

import lombok.SneakyThrows;

import java.io.*;
import java.util.function.Consumer;
import java.util.function.Predicate;
import java.util.zip.ZipEntry;
import java.util.zip.ZipInputStream;

public class ZipExtractor {

    public record HandledFile(String parent, byte[] contents) {
    }

    public record FileHandler(Predicate<String> shouldHandle, Consumer<HandledFile> handle) {
    }

    @SneakyThrows
    public static void extract(InputStream zippedInputStream, File destDir, FileHandler... fileHandlers) {
        byte[] buffer = new byte[1024];
        ZipInputStream zis = new ZipInputStream(zippedInputStream);
        ZipEntry zipEntry = zis.getNextEntry();

        while (zipEntry != null) {
            File newFile = newFile(destDir, zipEntry);
            if (zipEntry.isDirectory()) {
                if (!newFile.isDirectory() && !newFile.mkdirs()) {
                    throw new IOException("Failed to create directory " + newFile);
                }
            } else {
                // fix for Windows-created archives
                File parent = newFile.getParentFile();
                if (!parent.isDirectory() && !parent.mkdirs()) {
                    throw new IOException("Failed to create directory " + parent);
                }

                // write file content
                boolean someHandlerPredicateMatched = someHandlerPredicateMatched(newFile.getName(), fileHandlers);

                ByteArrayOutputStream fileContentsStream = someHandlerPredicateMatched ? new ByteArrayOutputStream() : null;
                try (FileOutputStream fos = new FileOutputStream(newFile)) {
                    int len;
                    while ((len = zis.read(buffer)) > 0) {
                        fos.write(buffer, 0, len);
                        if (someHandlerPredicateMatched) {
                            fileContentsStream.write(buffer, 0, len);
                        }
                    }
                }
                if (someHandlerPredicateMatched) {
                    for (FileHandler fileHandler : fileHandlers) {
                        if (fileHandler.shouldHandle().test(newFile.getName())) {
                            fileHandler.handle().accept(new HandledFile(zipEntry.getName(), fileContentsStream.toByteArray()));
                        }
                    }
                }
            }
            zipEntry = zis.getNextEntry();
        }
    }

    private static boolean someHandlerPredicateMatched(String name, FileHandler[] fileHandlers) {
        for (FileHandler fileHandler : fileHandlers) {
            if (fileHandler.shouldHandle.test(name)) {
                return true;
            }
        }
        return false;
    }

    private static File newFile(File destinationDir, ZipEntry zipEntry) throws IOException {
        File destFile = new File(destinationDir, zipEntry.getName());

        String destDirPath = destinationDir.getCanonicalPath();
        String destFilePath = destFile.getCanonicalPath();

        if (!destFilePath.startsWith(destDirPath + File.separator)) {
            throw new IOException("Entry is outside of the target dir: " + zipEntry.getName());
        }

        return destFile;
    }
}
