package yeetcd.controller.execution;

import java.io.File;

/**
 * Interface for mount input sources.
 * Supports both File-based (Docker) and PVC-based (Kubernetes) mounts.
 */
public interface MountInput {

    /**
     * Returns the directory for on-disk mounts (Docker execution).
     * For PVC-based mounts, this will throw UnsupportedOperationException.
     * Use instanceof to check mount type before calling.
     * 
     * @return the directory containing input files
     * @throws UnsupportedOperationException if this is a PVC-based mount
     */
    File directory();

    /**
     * Returns true if this is an on-disk mount (for Docker execution).
     */
    default boolean isOnDiskMount() {
        return this instanceof OnDiskMountInput;
    }

    /**
     * Returns true if this is a PVC-based mount (for Kubernetes execution).
     */
    default boolean isPvcMount() {
        return this instanceof PvcMountInput;
    }
}
