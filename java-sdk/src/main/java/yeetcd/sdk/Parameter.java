package yeetcd.sdk;

import yeetcd.protocol.pipeline.PipelineOuterClass;
import lombok.EqualsAndHashCode;
import lombok.ToString;

import java.util.Collections;
import java.util.List;

@EqualsAndHashCode
@ToString
public final class Parameter {

    private final TypeCheck typeCheck;
    private final boolean required;
    private final String defaultValue;
    private final List<String> choices;

    private Parameter(TypeCheck typeCheck, boolean required, String defaultValue, List<String> choices) {
        this.typeCheck = typeCheck;
        this.required = required;
        this.defaultValue = defaultValue;
        this.choices = choices;
    }

    public static Builder builder(TypeCheck typeCheck) {
        return new Builder(typeCheck);
    }

    public static class Builder {
        private final TypeCheck typeCheck;
        private boolean required = false;
        private String defaultValue = null;
        private List<String> choices = Collections.emptyList();

        public Builder(TypeCheck typeCheck) {
            this.typeCheck = typeCheck;
        }

        public Builder required(boolean required) {
            this.required = required;
            return this;
        }

        public Builder defaultValue(String defaultValue) {
            this.defaultValue = defaultValue;
            return this;
        }

        public Builder choices(List<String> choices) {
            this.choices = choices;
            return this;
        }

        public Parameter build() {
            return new Parameter(typeCheck, required, defaultValue, choices);
        }
    }

    PipelineOuterClass.Parameter toProtobuf() {
        PipelineOuterClass.Parameter.Builder builder = PipelineOuterClass.Parameter
            .newBuilder()
            .setTypeCheck(typeCheck.toProtobuf())
            .setRequired(required)
            .addAllChoices(choices == null ? Collections.emptyList() : choices);
        if (defaultValue != null) {
            builder = builder.setDefaultValue(defaultValue);
        }
        return builder.build();
    }

    public enum TypeCheck {
        STRING {
            @Override
            PipelineOuterClass.Parameter.TYPE_CHECK toProtobuf() {
                return PipelineOuterClass.Parameter.TYPE_CHECK.STRING;
            }
        },
        NUMBER {
            @Override
            PipelineOuterClass.Parameter.TYPE_CHECK toProtobuf() {
                return PipelineOuterClass.Parameter.TYPE_CHECK.NUMBER;
            }
        },
        BOOLEAN {
            @Override
            PipelineOuterClass.Parameter.TYPE_CHECK toProtobuf() {
                return PipelineOuterClass.Parameter.TYPE_CHECK.BOOLEAN;
            }
        };

        abstract PipelineOuterClass.Parameter.TYPE_CHECK toProtobuf();
    }
}
