package hooks

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/saireddy-shyamakura/springx/internal/postgen"
	"github.com/saireddy-shyamakura/springx/internal/prompt"
)

func init() {
	postgen.Register(&dockerHook{})
}

type dockerHook struct{}

func (d *dockerHook) Name() string { return "docker" }

func (d *dockerHook) Run(projectPath string, cfg *prompt.ProjectConfig) error {
	content := buildDockerfile(cfg)
	dest := filepath.Join(projectPath, "Dockerfile")
	return os.WriteFile(dest, []byte(content), 0o644)
}

// buildDockerfile produces a multi-stage, production-ready Dockerfile.
// It uses the official Eclipse Temurin images and the Spring Boot layered-jar
// feature to create minimal final images.
func buildDockerfile(cfg *prompt.ProjectConfig) string {
	javaVersion := cfg.JavaVersion
	if javaVersion == "" {
		javaVersion = "21"
	}

	baseImage := fmt.Sprintf("eclipse-temurin:%s-jre-jammy", javaVersion)
	builderImage := fmt.Sprintf("eclipse-temurin:%s-jdk-jammy", javaVersion)

	if isMavenProject(cfg.BuildTool) {
		return mavenDockerfile(builderImage, baseImage, cfg.ArtifactID)
	}
	return gradleDockerfile(builderImage, baseImage, cfg.ArtifactID)
}

func mavenDockerfile(builderImage, baseImage, artifactID string) string {
	return fmt.Sprintf(`# ── Build stage ──────────────────────────────────────────────────────────────
FROM %s AS builder
WORKDIR /workspace

# Cache Maven dependencies before copying full source.
COPY .mvn/ .mvn/
COPY mvnw pom.xml ./
RUN ./mvnw -q dependency:go-offline -B

COPY src ./src
RUN ./mvnw -q package -DskipTests -B

# Expand the Spring Boot fat-jar into layers for efficient image layering.
RUN java -Djarmode=layertools \
         -jar target/%s-*.jar \
         extract --destination target/extracted

# ── Runtime stage ─────────────────────────────────────────────────────────────
FROM %s AS runtime
WORKDIR /app

# Principle of least privilege: run as non-root.
RUN addgroup --system spring && adduser --system --ingroup spring spring
USER spring

# Copy Spring Boot layers in order of least→most frequently changed.
COPY --from=builder /workspace/target/extracted/dependencies/          ./
COPY --from=builder /workspace/target/extracted/spring-boot-loader/    ./
COPY --from=builder /workspace/target/extracted/snapshot-dependencies/ ./
COPY --from=builder /workspace/target/extracted/application/           ./

EXPOSE 8080

# Use exec form to ensure SIGTERM reaches the JVM.
ENTRYPOINT ["java", "org.springframework.boot.loader.launch.JarLauncher"]
`, builderImage, artifactID, baseImage)
}

func gradleDockerfile(builderImage, baseImage, artifactID string) string {
	return fmt.Sprintf(`# ── Build stage ──────────────────────────────────────────────────────────────
FROM %s AS builder
WORKDIR /workspace

# Cache Gradle wrapper and dependencies before copying full source.
COPY gradle/ gradle/
COPY gradlew settings.gradle* build.gradle* ./
RUN ./gradlew -q dependencies --no-daemon || true

COPY src ./src
RUN ./gradlew -q bootJar --no-daemon

# Expand the Spring Boot fat-jar into layers for efficient image layering.
RUN java -Djarmode=layertools \
         -jar build/libs/%s-*.jar \
         extract --destination build/extracted

# ── Runtime stage ─────────────────────────────────────────────────────────────
FROM %s AS runtime
WORKDIR /app

# Principle of least privilege: run as non-root.
RUN addgroup --system spring && adduser --system --ingroup spring spring
USER spring

# Copy Spring Boot layers in order of least→most frequently changed.
COPY --from=builder /workspace/build/extracted/dependencies/          ./
COPY --from=builder /workspace/build/extracted/spring-boot-loader/    ./
COPY --from=builder /workspace/build/extracted/snapshot-dependencies/ ./
COPY --from=builder /workspace/build/extracted/application/           ./

EXPOSE 8080

ENTRYPOINT ["java", "org.springframework.boot.loader.launch.JarLauncher"]
`, builderImage, artifactID, baseImage)
}

// isMavenProject returns true when the build tool ID refers to Maven.
func isMavenProject(buildTool string) bool {
	return strings.Contains(strings.ToLower(buildTool), "maven")
}
