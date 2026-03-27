package yeetcd.controller.pipeline;

import yeetcd.protocol.pipeline.PipelineOuterClass;

import java.util.List;

public record Parameter(TypeCheck typeCheck, boolean required, String defaultValue, List<String> choices) {

    void validateArgument(String argument) {
        typeCheck.validate(argument);
        if (choices != null && choices.size() > 0 && !choices.contains(argument)) {
            throw new IllegalArgumentException("Argument %s is not one of the allowed choices [%s]".formatted(
                    argument, String.join(", ", choices)
            ));
        }
    }

    public enum TypeCheck {
        STRING {
            @Override
            void validate(String input) {
            }
        },
        NUMBER {
            @Override
            void validate(String input) {
                //noinspection ResultOfMethodCallIgnored
                Double.parseDouble(input);
            }
        },
        BOOLEAN {
            @Override
            void validate(String input) {
                switch (input) {
                    case "true", "false" -> {}
                    default -> throw new IllegalArgumentException("Invalid boolean %s".formatted(input));
                }
            }
        };

        abstract void validate(String input);

        public static TypeCheck fromProtobuf(PipelineOuterClass.Parameter.TYPE_CHECK typeCheck) {
            switch (typeCheck) {
                case STRING -> {
                    return TypeCheck.STRING;
                }
                case NUMBER -> {
                    return TypeCheck.NUMBER;
                }
                case BOOLEAN -> {
                    return TypeCheck.BOOLEAN;
                }
                default -> throw new IllegalArgumentException("Unrecognised type check");
            }
        }

    }

    public static Parameter fromProtobuf(PipelineOuterClass.Parameter parameter) {
        return new Parameter(
                TypeCheck.fromProtobuf(parameter.getTypeCheck()),
                parameter.getRequired(),
                parameter.getDefaultValue(),
                parameter.getChoicesList().stream().toList()
        );
    }
}
