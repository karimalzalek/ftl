plugins {
  kotlin("jvm")
  id("java-gradle-plugin")
  id("com.squareup.wire") version "4.7.2"
}

group = "xyz.block"
version = "1.0-SNAPSHOT"

gradlePlugin {
  plugins {
    create("ftl") {
      id = "xyz.block.ftl"
      displayName = "FTL"
      implementationClass = "xyz.block.ftl.gradle.FTLPlugin"
      description = "Generate FTL stubs"
    }
  }
}

dependencies {
  compileOnly(gradleApi())
  implementation(project(":ftl-runtime"))

  // Use the Kotlin JUnit 5 integration.
  testImplementation(libs.kotlinTestJunit5)

  // Use the JUnit 5 integration.
  testImplementation(libs.junitJupiterEngine)
  testRuntimeOnly(libs.junitPlatformLauncher)

  implementation(libs.kotlinGradlePlugin)
  implementation(libs.kotlinPoet)
  implementation(libs.kotlinReflect)
  implementation(libs.kotlinxCoroutinesCore)
  implementation(libs.wireRuntime)
  implementation(libs.wireGrpcClient)
}

tasks.findByName("wrapper")?.enabled = false

wire {
  kotlin {
    rpcRole = "client"
    rpcCallStyle = "blocking"
  }
  sourcePath {
    srcDir("../../protos")
  }
}

tasks.named<Test>("test") {
  // Use JUnit Platform for unit tests.
  useJUnitPlatform()
  testLogging {
    events("passed", "skipped", "failed")
  }
}
