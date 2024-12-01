import 'package:flutter/material.dart';
import 'package:file_picker/file_picker.dart';
import 'package:http/http.dart' as http;
import 'dart:convert';
import 'package:http_parser/http_parser.dart';
import 'dart:io';
import 'package:dio/dio.dart';
import 'package:loading_animation_widget/loading_animation_widget.dart';

void main() {
  runApp(const MyApp());
}

class MyApp extends StatelessWidget {
  const MyApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'File Upload System',
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(seedColor: Colors.deepPurple),
        useMaterial3: true,
      ),
      debugShowCheckedModeBanner: false,
      home: const MyHomePage(title: 'File Upload System'),
    );
  }
}

class MyHomePage extends StatefulWidget {
  const MyHomePage({super.key, required this.title});

  final String title;

  @override
  State<MyHomePage> createState() => _MyHomePageState();
}

class _MyHomePageState extends State<MyHomePage> {
  final List<String> _fileList = [];
  String _uploadStatus = '';
  bool _isUploading = false;
  // Function to load the file list from the server
  Future<void> loadFiles() async {
    try {
      final response = await fetchFiles(); // Fetch files from API
      setState(() {
        _fileList.clear();
        _fileList.addAll(response); // Update file list
      });
    } catch (e) {
      setState(() {
        _uploadStatus = 'Error loading files. Please try again later.: $e';
      });
    }
  }

  Future<void> deleteFile(String file) async {
    var cookies = {'uuid': 'my-cookie'};
    var url = Uri.parse('https://example.com/api/delete/$file');
    try {
      // Make a GET request with cookies
      var response = await http.get(
        url,
        headers: {
          'Cookie': cookies.entries
              .map((entry) => '${entry.key}=${entry.value}')
              .join('; ') // Combine cookies into a single string
        },
      );

      if (response.statusCode == 200) {
        print('Request successful: ${response.body}');

        setState(() {
          _uploadStatus = 'Delete Sucesfull!';
        });

        loadFiles();
      } else {
        throw Exception('delete failed');
      }
    } catch (error) {
      print('Error: $error');
    }
  }

  Future<List<String>> fetchFiles() async {
    var url = Uri.parse('https://example.com/api/files');

    // Cookie value
    var cookies = {'uuid': 'my-cookie'};

    try {
      // Make a GET request with cookies
      var response = await http.get(
        url,
        headers: {
          'Cookie': cookies.entries
              .map((entry) => '${entry.key}=${entry.value}')
              .join('; ') // Combine cookies into a single string
        },
      );

      if (response.statusCode == 200) {
        print('Request successful: ${response.body}');

        var data = json.decode(response.body);

        // Extract the file names from the 'name' field of each object
        if (data is List) {
          return data
              .map<String>((file) => file['name'] as String)
              .toList(); // Extract 'name' from each object
        } else {
          throw Exception('Expected a list of files in the response');
        }
      } else {
        throw Exception('Request failed with status: ${response.statusCode}');
      }
    } catch (error) {
      print('Error: $error');
      return [];
    }
  }

  // Handle file upload
  Future<void> uploadFile(PlatformFile file) async {
    setState(() {
      _uploadStatus = 'Uploading...';
      _isUploading = true;
    });

    try {
      // Get the file from PlatformFile (this is for Flutter file picker plugin)
      File fileToUpload = File(file.path!);

      // URL to which the file will be uploaded
      var url = Uri.parse('https://example.com/api/upload');

      // Set the headers if you need a cookie
      var cookies = {'uuid': 'my-cookie'};
      var headers = {
        'Cookie': cookies.entries
            .map((entry) => '${entry.key}=${entry.value}')
            .join('; ') // Combine cookies into a single string
      };

      // Create a multipart request
      var request = http.MultipartRequest('POST', url)
        ..headers.addAll(headers)
        ..files.add(await http.MultipartFile.fromPath(
          'file',
          fileToUpload.path,
          contentType: MediaType('application', 'octet-stream'),
        ));

      // Send the request
      var response = await request.send();

      // Check if the response is successful
      if (response.statusCode == 200) {
        setState(() {
          _uploadStatus = 'File uploaded successfully!';
        });
        loadFiles(); // Refresh the file list after successful upload
      } else {
        setState(() {
          _uploadStatus = 'Upload failed: ${response.statusCode}';
        });
      }
    } catch (e) {
      setState(() {
        _uploadStatus = 'Upload failed: Network error';
      });
    } finally {
      setState(() {
        _isUploading = false; // Hide loading animation
      });
    }
  }

  // Function to download a file
  Future<void> downloadFile(String fileName) async {
    try {
      var url = 'https://example.com/api/download/$fileName';
      String directory;

      if (Platform.isLinux) {
        directory = '${Platform.environment['HOME']}/Desktop/transfer';
      } else if (Platform.isAndroid) {
        directory = '/storage/emulated/0/Download/transfer';
      } else {
        throw Exception('Unsupported platform');
      }

      final filePath = '$directory/$fileName';
      await Directory(directory).create(recursive: true);

      // Create Dio instance with cookie
      final dio = Dio();
      dio.options.headers['cookie'] = 'uuid=my-cookie';

      await dio.download(url, filePath);

      setState(() {
        _uploadStatus = 'File downloaded successfully to $filePath';
      });
    } catch (e) {
      setState(() {
        _uploadStatus = 'Download failed: $e';
      });
    }
  }

  @override
  void initState() {
    super.initState();
    loadFiles();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text(widget.title),
        backgroundColor: Colors.deepPurple,
        actions: [
          IconButton(
            icon: Icon(Icons.refresh),
            onPressed: loadFiles, // Refresh the file list on button click
          ),
        ],
      ),
      body: Padding(
        padding: const EdgeInsets.all(20.0),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // File upload form
            Container(
              padding: const EdgeInsets.all(20),
              decoration: BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.circular(8),
                boxShadow: [
                  BoxShadow(
                    color: Colors.black.withOpacity(0.1),
                    blurRadius: 4,
                  ),
                ],
              ),
              child: Column(
                children: [
                  const Text('Upload File', style: TextStyle(fontSize: 18)),
                  const SizedBox(height: 10),
                  ElevatedButton(
                    onPressed: () async {
                      FilePickerResult? result =
                          await FilePicker.platform.pickFiles();
                      if (result != null) {
                        PlatformFile file = result.files.first;
                        await uploadFile(file); // Upload the selected file
                      }
                    },
                    child: const Text('Choose file'),
                  ),
                  const SizedBox(height: 10),
                  Text(_uploadStatus),
                ],
              ),
            ),
            const SizedBox(height: 20),
            if (_isUploading)
              Center(
                child: LoadingAnimationWidget.halfTriangleDot(
                  color: const Color(0xFFFF69B4), // Bright Pink
                  size: 200,
                ),
              ),
            // Scrollable file list
            Expanded(
              child: SingleChildScrollView(
                child: Container(
                  padding: const EdgeInsets.all(20),
                  decoration: BoxDecoration(
                    color: Colors.white,
                    borderRadius: BorderRadius.circular(8),
                    boxShadow: [
                      BoxShadow(
                        color: Colors.black.withOpacity(0.1),
                        blurRadius: 4,
                      ),
                    ],
                  ),
                  child: Column(
                    children: [
                      const Text('Available Files',
                          style: TextStyle(fontSize: 18)),
                      const SizedBox(height: 10),
                      ..._fileList.map((file) {
                        return ListTile(
                          title: Text(file),
                          trailing: Row(
                            mainAxisSize: MainAxisSize.min,
                            children: [
                              TextButton(
                                onPressed: () {
                                  deleteFile(file);
                                },
                                child: const Text(
                                  'Delete',
                                  style: TextStyle(color: Colors.red),
                                ),
                              ),
                              TextButton(
                                onPressed: () {
                                  downloadFile(file);
                                },
                                child: const Text(
                                  'Download',
                                  style: TextStyle(color: Colors.blue),
                                ),
                              ),
                            ],
                          ),
                        );
                      }),
                    ],
                  ),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
