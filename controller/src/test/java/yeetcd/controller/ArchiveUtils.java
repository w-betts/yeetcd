package yeetcd.controller;

import lombok.SneakyThrows;

import java.io.ByteArrayOutputStream;
import java.io.File;
import java.io.FileInputStream;
import java.io.IOException;
import java.util.Map;
import java.util.zip.ZipEntry;
import java.util.zip.ZipOutputStream;

public class ArchiveUtils {
    private static final byte[] projectZip = initialiseProjectZip();

    @SneakyThrows
    private static byte[] initialiseProjectZip()  {
        File fileToZip = new File(System.getProperty("user.dir").replaceAll("controller/?", "/"));
        ByteArrayOutputStream outputStream = new ByteArrayOutputStream();
        try (ZipOutputStream zipOut = new ZipOutputStream(outputStream)) {
            ArchiveUtils.zipFile(fileToZip, fileToZip.getName(), zipOut);
        }
        return outputStream.toByteArray();
    }

    @SneakyThrows
    public static byte[] projectZip()  {
        return projectZip;
    }

    @SneakyThrows
    public static byte[] createZip(Map<String, byte[]> pathContents) {
        ByteArrayOutputStream outputStream = new ByteArrayOutputStream();
        try (ZipOutputStream zipOut = new ZipOutputStream(outputStream)) {
            for (Map.Entry<String, byte[]> entry : pathContents.entrySet()) {
                String path = entry.getKey();
                zipOut.putNextEntry(new ZipEntry(path));
                zipOut.write(entry.getValue());
            }
        }
        return outputStream.toByteArray();
    }

    public static void zipFile(File fileToZip, String fileName, ZipOutputStream zipOut) throws IOException {
        if (fileToZip.isHidden()) {
            return;
        }
        if (fileToZip.isDirectory() && fileToZip.getName().equals("target")) {
            return;
        }
        if (fileToZip.isDirectory()) {
            if (fileName.endsWith("/")) {
                zipOut.putNextEntry(new ZipEntry(fileName));
                zipOut.closeEntry();
            } else {
                zipOut.putNextEntry(new ZipEntry(fileName + "/"));
                zipOut.closeEntry();
            }
            File[] children = fileToZip.listFiles();
            for (File childFile : children) {
                zipFile(childFile, fileName + "/" + childFile.getName(), zipOut);
            }
            return;
        }
        FileInputStream fis = new FileInputStream(fileToZip);
        ZipEntry zipEntry = new ZipEntry(fileName);
        zipOut.putNextEntry(zipEntry);
        byte[] bytes = new byte[1024];
        int length;
        while ((length = fis.read(bytes)) >= 0) {
            zipOut.write(bytes, 0, length);
        }
        fis.close();
    }
}
