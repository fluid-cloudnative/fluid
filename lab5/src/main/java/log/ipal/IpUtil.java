package log.ipal;

import org.lionsoul.ip2region.DataBlock;
import org.lionsoul.ip2region.DbConfig;
import org.lionsoul.ip2region.DbSearcher;
import org.lionsoul.ip2region.Util;

import java.io.File;
import java.io.InputStream;
import java.lang.reflect.Method;
import java.nio.charset.StandardCharsets;
import java.util.Objects;
import java.util.Scanner;

public class IpUtil {
    public static String getCityInfo(String ip){
//db
        String dbPath = Objects.requireNonNull(IpUtil.class.getResource("/ip2region.db")).getPath();
       //String dbPath="/home/orangecattt/Code/hadoop/lab5/src/main/resources/ip2region.db";

        File file = new File("dbPath");
        if ( file.exists() == false ) {
           /* System.out.println("Error: Invalid ip2region.db file");*/
        }
//查询算法
        int algorithm = DbSearcher.BTREE_ALGORITHM; //B-tree
//DbSearcher.BINARY_ALGORITHM //Binary
//DbSearcher.MEMORY_ALGORITYM //Memory
        try {
            DbConfig config = new DbConfig();
            DbSearcher searcher = new DbSearcher(config, dbPath);
//define the method
            Method method = null;
            switch ( algorithm )
            {
                case DbSearcher.BTREE_ALGORITHM:
                    method = searcher.getClass().getMethod("btreeSearch", String.class);
                    break;
                case DbSearcher.BINARY_ALGORITHM:
                    method = searcher.getClass().getMethod("binarySearch", String.class);
                    break;
                case DbSearcher.MEMORY_ALGORITYM:
                    method = searcher.getClass().getMethod("memorySearch", String.class);
                    break;
            }
            DataBlock dataBlock = null;
            if ( Util.isIpAddress(ip) == false ) {
               /* System.out.println("Error: Invalid ip address");*/
            }
           /* System.out.println(ip);*/
            dataBlock = (DataBlock) method.invoke(searcher, ip);
            return dataBlock.getRegion();
        } catch (Exception e) {
            e.printStackTrace();
        }
        return null;
    }


}