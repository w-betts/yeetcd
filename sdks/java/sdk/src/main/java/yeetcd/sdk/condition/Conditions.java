package yeetcd.sdk.condition;

public final class Conditions {

    public static Condition not(Condition condition) {
        return NotCondition.builder(condition).build();
    }

    public static Condition and(Condition left, Condition right) {
        return AndCondition.builder(left, right).build();
    }

    public static Condition or(Condition left, Condition right) {
        return OrCondition.builder(left, right).build();
    }

    public static Condition workContextCondition(String key, WorkContextCondition.Operand operand, String value) {
        return WorkContextCondition.builder(key, operand, value).build();
    }

    public static Condition previousWorkStatusCondition(PreviousWorkStatusCondition.Status status) {
        return PreviousWorkStatusCondition.builder(status).build();
    }
}
