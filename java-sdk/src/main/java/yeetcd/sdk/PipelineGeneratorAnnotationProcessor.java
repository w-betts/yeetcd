package yeetcd.sdk;

import com.google.auto.service.AutoService;

import javax.annotation.processing.*;
import javax.lang.model.SourceVersion;
import javax.lang.model.element.Element;
import javax.lang.model.element.PackageElement;
import javax.lang.model.element.TypeElement;
import javax.tools.JavaFileObject;
import java.io.IOException;
import java.io.OutputStreamWriter;
import java.io.PrintWriter;
import java.util.Set;
import java.util.stream.Collectors;

@SupportedAnnotationTypes("yeetcd.sdk.PipelineGenerator")
@SupportedSourceVersion(SourceVersion.RELEASE_17)
@AutoService(Processor.class)
public final class PipelineGeneratorAnnotationProcessor extends AbstractProcessor {

    @Override
    public boolean process(Set<? extends TypeElement> annotations, RoundEnvironment roundEnv) {
        String commaSeparatedPipelines = annotations
                .stream()
                .flatMap(element -> roundEnv.getElementsAnnotatedWith(element).stream())
                .map(PipelineGeneratorAnnotationProcessor::pipelineGeneratorInvocationSnippet)
                .collect(Collectors.joining(", "));
        writePipelineDefinitionsClass(commaSeparatedPipelines);
        writeCustomWorkRunnerClass(commaSeparatedPipelines);
        return true;
    }

    private void writeCustomWorkRunnerClass(String commaSeparatedPipelines) {
        try {
            JavaFileObject pipelineDefinitionsMainClass = processingEnv.getFiler().createSourceFile("yeetcd.sdk.GeneratedCustomWorkRunner");
            try (PrintWriter writer = new PrintWriter(new OutputStreamWriter(pipelineDefinitionsMainClass.openOutputStream()))) {
                writer.println("""
                        package yeetcd.sdk;
                        
                        import java.io.IOException;
                                                
                        public class GeneratedCustomWorkRunner {
                            public static void main(String[] args) throws IOException {
                                // Initialise all the pipelines to fill up the static map
                                new Pipelines(%s);
                                Pipeline.runNativeWorkDefinition(args[0], args[1]);
                            }
                        }
                        """.formatted(commaSeparatedPipelines));
            }
        } catch (IOException ex) {
            ex.printStackTrace();
        }
    }

    private void writePipelineDefinitionsClass(String commaSeparatedPipelines) {
        try {
            JavaFileObject pipelineDefinitionsMainClass = processingEnv.getFiler().createSourceFile("yeetcd.sdk.GeneratedPipelineDefinitions");
            try (PrintWriter writer = new PrintWriter(new OutputStreamWriter(pipelineDefinitionsMainClass.openOutputStream()))) {
                writer.println("""
                        package yeetcd.sdk;
                        
                        import java.io.IOException;
                                                
                        public class GeneratedPipelineDefinitions {
                            public static void main(String[] args) throws IOException {
                                new Pipelines(%s).toProtobuf().writeTo(System.out);
                                System.out.flush();
                            }
                        }
                        """.formatted(commaSeparatedPipelines));
            }
        } catch (IOException ex) {
            ex.printStackTrace();
        }
    }

    private static String pipelineGeneratorInvocationSnippet(Element element) {
        StringBuilder stringBuilder = new StringBuilder();
        Element enclosingElement = element;
        while ((enclosingElement = enclosingElement.getEnclosingElement()) != null) {
            String name = enclosingElement instanceof PackageElement ?
                    ((PackageElement) enclosingElement).getQualifiedName().toString() :
                    enclosingElement.getSimpleName().toString();
            if (!name.isBlank()) {
                stringBuilder.insert(0, name + ".");
            }
        }
        stringBuilder
                .append(element.getSimpleName().toString())
                .append("()");
        return stringBuilder.toString();
    }
}
