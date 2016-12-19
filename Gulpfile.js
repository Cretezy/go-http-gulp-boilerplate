const gulp = require("gulp");
const util = require("gulp-util");
const sass = require("gulp-sass");
const sourcemaps = require('gulp-sourcemaps');
const autoprefixer = require('gulp-autoprefixer');
const concat = require("gulp-concat");
const notifier = require("node-notifier");
const sync = require("gulp-sync")(gulp).sync;
const child = require("child_process");
const os = require("os");
const path = require('path');
const minifyCss = require("gulp-minify-css");
const webpack = require('webpack');
const gulpWebpack = require('gulp-webpack');

if (util.env.production) {
    process.env.NODE_ENV = "production";
}

const windows = os.platform() == "win32";
const path_folder = __dirname.split(windows ? "\\" : "/");
const name = path_folder[path_folder.length - 1];

const config = {
    name: name,
    production: process.env.NODE_ENV == "production",
};

logger("config", "green")("Using config: " + JSON.stringify(config));

let server = null;

gulp.task("scripts:build", () => {
    let webpackConfig = {
        devtool: config.production ? 'cheap-module-source-map' : 'cheap-module-eval-source-map',
        output: {
            filename: 'index.js',
        },
        module: {
            loaders: [
                {
                    loader: "babel-loader",
                    include: [
                        path.resolve(__dirname, "static/scripts"),
                    ],
                    test: /\.js$/,

                    query: {
                        presets: ['es2015'],
                        plugins: ['transform-runtime']
                    }
                },
            ]
        },
        plugins: [
            new webpack.DefinePlugin({
                'process.env': {
                    'NODE_ENV': config.production ? JSON.stringify('production') : process.env.NODE_ENV
                }
            })
        ],
    };
    if (config.production) {
        webpackConfig.plugins.push(new webpack.optimize.UglifyJsPlugin({sourceMap: false}))
    }

    gulp.src("static/scripts/main.js")
        .pipe(gulpWebpack(webpackConfig, webpack, function (err, stats) {
            if (err) {
                logger("scripts", "red")(err);
            } else {
                logger("scripts", "yellow")(stats.toString({colors: true}));
            }
        }))
        .pipe(gulp.dest("public/build"));
});

gulp.task("scripts:watch", () => {
    gulp.watch([
        "static/scripts/**/*.js",
    ], sync([
        "scripts:build",
    ], "scripts"));
});

gulp.task("styles:build", () => {
    gulp.src("static/styles/main.scss")
        .pipe(config.production ? util.noop() : sourcemaps.init())
        .pipe(sass().on("error", (err) => {
            logger("styles", "red")(err.message)
        }))
        .pipe(concat("index.css"))
        .pipe(!config.production ? util.noop() : minifyCss())
        .pipe(config.production ? util.noop() : sourcemaps.write())
        .pipe(autoprefixer())
        .pipe(gulp.dest("public/build"));
});

gulp.task("styles:watch", () => {
    gulp.watch([
        "static/styles/**/*.scss",
    ], sync([
        "styles:build",
    ], "styles"));
});

gulp.task("server:build", () => {
    // Build
    const bindata = child.spawnSync("go-bindata", [!config.production ? "-debug" : "", "public/...", "views/..."]);
    const build = child.spawnSync("go", !config.production ? ["install"] : ["build", "-o", "build/" + config.name]);

    // Compilation error
    if (build.stderr.length) {
        util.log(util.colors.red("Error: server:build"));

        const lines = build.stderr.toString()
            .split("\n").filter((line) => {
                return line.length
            });

        lines.forEach((line) => {
            util.log(util.colors.yellow(line));
        });

        notifier.notify({
            title: "Error: server:build",
            message: lines.join("\n")
        });
    }

    return build;
});

gulp.task("server:run", () => {
    // Stop the server
    if (server) {
        server.kill();
    }

    // Run the server
    //server = child.spawn(path.join(__dirname, app + (windows ? ".exe" : "")));
    server = child.spawn(config.name + (windows ? ".exe" : ""));

    // Display output
    server.stderr.on("data", logger("server", "red"));
    server.stdout.on("data", logger("server", "yellow"));
});

function logger(prefix, color) {
    return (data) => {
        data.toString().split("\n").filter((line) => {
            return line.length
        }).forEach((line) => {
            util.log(util.colors[color]("[" + prefix + "]"), line + util.colors.reset(" "));
        });
    }
}

// Watch files
gulp.task("server:watch", () => {
    gulp.watch([
        "**/*.go",
        "views/**/*.tmpl",
        "!bindata.go"
    ], sync([
        "server:build",
        "server:run"
    ], "server"));
});

gulp.task("default", ["build", "watch", "run"]);

gulp.task("watch", ["scripts:watch", "styles:watch", "server:watch"]);

gulp.task("build", sync(["scripts:build", "styles:build", "server:build"], "build"));

gulp.task("run", ["server:run"]);

gulp.on('end', () => {
    // Stop the server
    if (server) {
        server.kill();
    }
});
