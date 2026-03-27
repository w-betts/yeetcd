package yeetcd.controller.execution;

import com.google.common.annotations.VisibleForTesting;
import org.apache.commons.io.output.TeeOutputStream;

import java.io.*;

public class JobStreams {
    private final ByteArrayOutputStream stdOutByteArrayOutputStream = new ByteArrayOutputStream();
    private final ByteArrayOutputStream stdErrByteArrayOutputStream = new ByteArrayOutputStream();
    private OutputStream stdOutOutputStream;
    private OutputStream stdErrOutputStream;

    @VisibleForTesting
    protected JobStreams() {
    }

    public JobStreams(OutputStream stdOutOutputStream, OutputStream stdErrOutputStream) {
        this.stdOutOutputStream = new TeeOutputStream(stdOutOutputStream, stdOutByteArrayOutputStream);
        this.stdErrOutputStream = new TeeOutputStream(stdErrOutputStream, stdErrByteArrayOutputStream);
    }

    public OutputStream getStdOutOutputStream() {
        return stdOutOutputStream;
    }

    public OutputStream getStdErrOutputStream() {
        return stdErrOutputStream;
    }

    public byte[] getStdOut() {
        return stdOutByteArrayOutputStream.toByteArray();
    }

    public byte[] getStdErr() {
        return stdErrByteArrayOutputStream.toByteArray();
    }
}
