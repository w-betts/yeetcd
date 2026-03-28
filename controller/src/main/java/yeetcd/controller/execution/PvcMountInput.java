package yeetcd.controller.execution;

/**
 * MountInput implementation for PVC-based mounts.
 * Used by KubernetesExecutionEngine to mount S3-backed PVCs into pods.
 */
public class PvcMountInput implements MountInput {

    private final String pvcName;
    private final String subPath;

    public PvcMountInput(String pvcName, String subPath) {
        this.pvcName = pvcName;
        this.subPath = subPath;
    }

    public String pvcName() {
        return pvcName;
    }

    public String subPath() {
        return subPath;
    }

    @Override
    public java.io.File directory() {
        throw new UnsupportedOperationException("PvcMountInput does not support direct file access. Use pvcName() and subPath() instead.");
    }
}
