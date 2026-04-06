package yeetcd.test;

import java.util.Collections;
import java.util.List;
import java.util.Map;

public record WorkExecution(
    WorkBehaviorType type,
    String matchKey,
    String image,
    String[] cmd,
    Map<String, String> envVars,
    String workingDir,
    int exitCode,
    String stdout,
    String stderr
) {
    public WorkExecution {
        if (envVars == null) envVars = Collections.emptyMap();
    }

    public static WorkExecution containerised(String image, String[] cmd, Map<String, String> envVars, String workingDir, int exitCode, String stdout, String stderr) {
        return new WorkExecution(WorkBehaviorType.CONTAINERISED, image, image, cmd, envVars, workingDir, exitCode, stdout, stderr);
    }

    public static WorkExecution custom(String executionId, String image, String[] cmd, Map<String, String> envVars, String workingDir, int exitCode, String stdout, String stderr) {
        return new WorkExecution(WorkBehaviorType.CUSTOM, executionId, image, cmd, envVars, workingDir, exitCode, stdout, stderr);
    }

    public static WorkExecution dynamic(String image, String[] cmd, Map<String, String> envVars, String workingDir, int exitCode, String stdout, String stderr) {
        return new WorkExecution(WorkBehaviorType.DYNAMIC, null, image, cmd, envVars, workingDir, exitCode, stdout, stderr);
    }
}
