package io.fluid.demo;

import org.apache.hadoop.conf.Configuration;
import org.apache.hadoop.fs.*;

import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStreamReader;
import java.net.URI;

public class HDFSClient {

    private static final String HDFS_URL = "alluxio://hadoop-master-0.default.svc.cluster.local:"+ System.getenv("HADOOP_PORT") + "/hadoop";
    private static final String FILE_PATH = "/hadoop/RELEASENOTES.md";
    private static final String BASE_PATH = "/hadoop/";

    public static void main(String[] args) throws IOException {
        Configuration conf = new Configuration();
        FileSystem fs = FileSystem.get(URI.create(HDFS_URL), conf);

        readFile(fs, FILE_PATH);

        listFiles(fs, new Path(BASE_PATH));

        long start = System.currentTimeMillis();
        copyDirectory(fs, BASE_PATH, "./");
        long cost = System.currentTimeMillis() - start;
        System.out.println("copy directory cost:" + cost + "ms");
    }

    private static void readFile(FileSystem fs, String file) throws IOException {
        FSDataInputStream is = fs.open(new Path(file));
        BufferedReader reader = new BufferedReader(new InputStreamReader(is));
        String line;
        while((line = reader.readLine()) != null) {
            System.out.println(line);
        }
    }

    private static void listFiles(FileSystem fs, Path path) throws IOException {
        FileStatus[] files = fs.listStatus(path);
        for(FileStatus file : files) {
            if(file.isDirectory()) {
                listFiles(fs, file.getPath());
            } else {
                System.out.println("## " + file.getPath().getName());
            }
        }
    }

    public static void copyDirectory(FileSystem fs, String sourceDir, String targetDir) throws IOException {
        Path source = new Path(sourceDir);
        RemoteIterator<LocatedFileStatus> sourceFiles = fs.listFiles(source, true);
        if(sourceFiles != null) {
            while(sourceFiles.hasNext()){
                Path nextPath = sourceFiles.next().getPath();
                String targetFile = targetDir + nextPath.getName();
                fs.copyToLocalFile(nextPath, new Path(targetFile));
            }
        }
    }
}
