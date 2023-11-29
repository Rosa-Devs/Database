width = window.innerWidth;
height = window.innerHeight;
fetch('/tree')
      .then(response => response.json())
      .then(graphData => {
        // Process the graph data as needed

        // Extract unique nodes from links
        const nodeIds = Array.from(new Set(graphData.links.flatMap(link => [link.source, link.target])));
        const nodesData = nodeIds.map(id => ({ id }));

        // Ensure that graphData.links is an array
        const linksData = Array.isArray(graphData.links) ? graphData.links : [];

        // Create an SVG container
        const svg = d3.select("body").append("svg")
          .attr("width", width)
          .attr("height", height)
          .append("g");

        // Create a force simulation
        const simulation = d3.forceSimulation(nodesData)
          .force("link", d3.forceLink(linksData).id(d => d.id))
          .force("charge", d3.forceManyBody().strength(-200))
          .force("center", d3.forceCenter(width - (width/2), height - (height/2)));

        // Create links
        const links = svg.selectAll(".link")
          .data(linksData)
          .enter().append("line")
          .attr("class", "link")
          .style("stroke", "gray")
          .style("stroke-width", 2);

        // Create nodes
        const nodes = svg.selectAll(".node")
          .data(nodesData)
          .enter().append("circle")
          .attr("class", "node")
          .attr("r", 10)
          .style("fill", "black");

        // Update link positions in each tick
        simulation.nodes(nodesData)
          .on("tick", tick);

        // Update link positions in each tick
        function tick() {
          links
            .attr("x1", d => d.source.x)
            .attr("y1", d => d.source.y)
            .attr("x2", d => d.target.x)
            .attr("y2", d => d.target.y);

          nodes
            .attr("cx", d => d.x)
            .attr("cy", d => d.y);
        }

        const zoom = d3.zoom()
          .scaleExtent([0.1, 3]) // Adjust the scale extent as needed
          .on("zoom", zoomed);

        svg.call(zoom);

        // Function to handle zooming
        function zoomed(event) {
          svg.attr("transform", event.transform);
        }

          // Add mouse drag behavior for nodes
        nodes.call(d3.drag()
        .on("start", dragstarted)
        .on("drag", dragged)
        .on("end", dragended));

        // Function to handle drag start
        function dragstarted(event, d) {
          if (!event.active) simulation.alphaTarget(0.3).restart();
          d.fx = d.x;
          d.fy = d.y;
        }

        // Function to handle dragging
        function dragged(event, d) {
          d.fx = event.x;
          d.fy = event.y;
        }

        // Function to handle drag end
        function dragended(event, d) {
          if (!event.active) simulation.alphaTarget(0);
            d.fx = null;
            d.fy = null;
          }

       

        // Function to handle window resize
        
      })
      .catch(error => console.error('Error fetching graph data:', error));

window.addEventListener("resize", function () {
  width = window.innerWidth;
  height = window.innerHeight;
  force.size([width, height]).start();
  svg.attr("width", width).attr("height", height);
});