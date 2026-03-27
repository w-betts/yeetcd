package yeetcd.controller.pipeline;

import org.junit.jupiter.api.Test;

import java.util.Collections;
import java.util.List;
import java.util.Map;

import static org.hamcrest.CoreMatchers.equalTo;
import static org.hamcrest.MatcherAssert.assertThat;
import static org.junit.jupiter.api.Assertions.assertThrows;

public class ArgumentsTest {

    @Test
    public void shouldPassValidationForStringWhichIsValidChoice() {
        // given
        String paramName = "param";
        String argumentValue = "value1";
        Parameters parameters = new Parameters(Map.of(
                paramName, new Parameter(Parameter.TypeCheck.STRING, true, null, List.of(argumentValue, "value2"))
        ));

        Arguments arguments = new Arguments(Map.of(
                paramName, argumentValue
        ));

        // when
        WorkContext validatedWorkContext = arguments.asValidatedWorkContext(parameters);

        // then
        assertThat(validatedWorkContext.workContextMap(), equalTo(Map.of(paramName, argumentValue)));
    }

    @Test
    public void shouldPassValidationForStringWhenNoChoicesDefined() {
        // given
        String paramName = "param";
        String argumentValue = "value1";
        Parameters parameters = new Parameters(Map.of(
                paramName, new Parameter(Parameter.TypeCheck.STRING, true, null, Collections.emptyList())
        ));

        Arguments arguments = new Arguments(Map.of(
                paramName, argumentValue
        ));

        // when
        WorkContext validatedWorkContext = arguments.asValidatedWorkContext(parameters);

        // then
        assertThat(validatedWorkContext.workContextMap(), equalTo(Map.of(paramName, argumentValue)));
    }

    @Test
    public void shouldPassValidationForNumber() {
        // given
        String paramName = "param";
        String argumentValue = "1";
        Parameters parameters = new Parameters(Map.of(
                paramName, new Parameter(Parameter.TypeCheck.NUMBER, true, null, Collections.emptyList())
        ));

        Arguments arguments = new Arguments(Map.of(
                paramName, argumentValue
        ));

        // when
        WorkContext validatedWorkContext = arguments.asValidatedWorkContext(parameters);

        // then
        assertThat(validatedWorkContext.workContextMap(), equalTo(Map.of(paramName, argumentValue)));
    }

    @Test
    public void shouldPassValidationForBoolean() {
        // given
        String paramName = "param";
        String argumentValue = "false";
        Parameters parameters = new Parameters(Map.of(
                paramName, new Parameter(Parameter.TypeCheck.BOOLEAN, true, null, Collections.emptyList())
        ));

        Arguments arguments = new Arguments(Map.of(
                paramName, argumentValue
        ));

        // when
        WorkContext validatedWorkContext = arguments.asValidatedWorkContext(parameters);

        // then
        assertThat(validatedWorkContext.workContextMap(), equalTo(Map.of(paramName, argumentValue)));
    }

    @Test
    public void shouldFailValidationWhenNotValidNumber() {
        // given
        String paramName = "param";
        String argumentValue = "value1";
        Parameters parameters = new Parameters(Map.of(
                paramName, new Parameter(Parameter.TypeCheck.NUMBER, true, null, null)
        ));

        Arguments arguments = new Arguments(Map.of(
                paramName, argumentValue
        ));

        // when / then
        assertThrows(IllegalArgumentException.class, () -> arguments.asValidatedWorkContext(parameters));
    }

    @Test
    public void shouldFailValidationWhenNotValidBoolean() {
        // given
        String paramName = "param";
        String argumentValue = "value1";
        Parameters parameters = new Parameters(Map.of(
                paramName, new Parameter(Parameter.TypeCheck.BOOLEAN, true, null, null)
        ));

        Arguments arguments = new Arguments(Map.of(
                paramName, argumentValue
        ));

        // when / then
        assertThrows(IllegalArgumentException.class, () -> arguments.asValidatedWorkContext(parameters));
    }

    @Test
    public void shouldFailValidationWhenNotAnAllowedChoice() {
        // given
        String paramName = "param";
        String argumentValue = "value1";
        Parameters parameters = new Parameters(Map.of(
                paramName, new Parameter(Parameter.TypeCheck.STRING, true, null, List.of("other"))
        ));

        Arguments arguments = new Arguments(Map.of(
                paramName, argumentValue
        ));

        // when / then
        assertThrows(IllegalArgumentException.class, () -> arguments.asValidatedWorkContext(parameters));
    }

    @Test
    public void shouldFailValidationWhenRequiredArgumentNotPresent() {
        // given
        String paramName = "param";
        Parameters parameters = new Parameters(Map.of(
                paramName, new Parameter(Parameter.TypeCheck.STRING, true, null, List.of("other"))
        ));

        Arguments arguments = new Arguments(Map.of());

        // when / then
        assertThrows(IllegalArgumentException.class, () -> arguments.asValidatedWorkContext(parameters));
    }
}
